package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotParticipant       = errors.New("not a participant in this conversation")
)

type ConversationRepo struct {
	pool *pgxpool.Pool
}

func NewConversationRepo(pool *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{pool: pool}
}

func (r *ConversationRepo) FindOrCreate(ctx context.Context, buyerID, sellerID string, listingID *string) (*model.Conversation, error) {
	var conv model.Conversation

	// Try to find existing conversation between these two users (either direction)
	// Uses idx_conversations_participants index: LEAST/GREATEST
	err := r.pool.QueryRow(ctx,
		`SELECT id, member_id, seller_id, listing_id, last_message_at, created_at
		 FROM conversations
		 WHERE LEAST(member_id, seller_id) = LEAST($1::uuid, $2::uuid)
		   AND GREATEST(member_id, seller_id) = GREATEST($1::uuid, $2::uuid)
		 ORDER BY last_message_at DESC LIMIT 1`,
		buyerID, sellerID,
	).Scan(&conv.ID, &conv.MemberID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
	if err == nil {
		return &conv, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Create new
	err = r.pool.QueryRow(ctx,
		`INSERT INTO conversations (member_id, seller_id, listing_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, member_id, seller_id, listing_id, last_message_at, created_at`,
		buyerID, sellerID, listingID,
	).Scan(&conv.ID, &conv.MemberID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) GetByID(ctx context.Context, id string) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.pool.QueryRow(ctx,
		`SELECT id, member_id, seller_id, listing_id, last_message_at, created_at
		 FROM conversations WHERE id = $1`,
		id,
	).Scan(&conv.ID, &conv.MemberID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.Conversation, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM conversations WHERE member_id = $1 OR seller_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.member_id, c.seller_id, c.listing_id, c.last_message_at, c.created_at,
		        u.id, u.role, u.name, u.avatar_url, u.province, u.description, u.org_name, u.created_at,
		        COALESCE(unread.cnt, 0)
		 FROM conversations c
		 JOIN users u ON u.id = CASE WHEN c.member_id = $1 THEN c.seller_id ELSE c.member_id END
		 LEFT JOIN (
		     SELECT conversation_id, COUNT(*) AS cnt
		     FROM messages
		     WHERE sender_id <> $1 AND read_at IS NULL
		     GROUP BY conversation_id
		 ) unread ON unread.conversation_id = c.id
		 WHERE c.member_id = $1 OR c.seller_id = $1
		 ORDER BY c.last_message_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var convs []*model.Conversation
	for rows.Next() {
		var conv model.Conversation
		conv.OtherUser = &model.PublicProfile{}
		if err := rows.Scan(
			&conv.ID, &conv.MemberID, &conv.SellerID, &conv.ListingID,
			&conv.LastMessageAt, &conv.CreatedAt,
			&conv.OtherUser.ID, &conv.OtherUser.Role, &conv.OtherUser.Name,
			&conv.OtherUser.AvatarURL, &conv.OtherUser.Province,
			&conv.OtherUser.Description, &conv.OtherUser.OrgName, &conv.OtherUser.CreatedAt,
			&conv.UnreadCount,
		); err != nil {
			return nil, 0, err
		}
		convs = append(convs, &conv)
	}
	if convs == nil {
		convs = []*model.Conversation{}
	}
	return convs, total, rows.Err()
}

func (r *ConversationRepo) SendMessage(ctx context.Context, conversationID, senderID, content, msgType string, replyToID *string) (*model.Message, error) {
	if msgType == "" {
		msgType = "text"
	}

	var msg model.Message
	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, sender_id, content, type, reply_to_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, conversation_id, sender_id, content, type, read_at, reply_to_id, created_at`,
		conversationID, senderID, content, msgType, replyToID,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Type, &msg.ReadAt, &msg.ReplyToID, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Load reply_to message if present
	if msg.ReplyToID != nil {
		r.loadReplyTo(ctx, &msg)
	}

	// Update last_message_at
	_, _ = r.pool.Exec(ctx,
		`UPDATE conversations SET last_message_at = NOW() WHERE id = $1`, conversationID,
	)

	return &msg, nil
}

func (r *ConversationRepo) loadReplyTo(ctx context.Context, msg *model.Message) {
	if msg.ReplyToID == nil {
		return
	}
	var reply model.ReplyMessage
	err := r.pool.QueryRow(ctx,
		`SELECT id, sender_id, content, type FROM messages WHERE id = $1`, *msg.ReplyToID,
	).Scan(&reply.ID, &reply.SenderID, &reply.Content, &reply.Type)
	if err == nil {
		msg.ReplyTo = &reply
	}
}

func (r *ConversationRepo) GetMessages(ctx context.Context, conversationID, readerID string, page, limit int) ([]*model.Message, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages
		 WHERE conversation_id = $1
		   AND NOT (sender_id = $2 AND deleted_by_sender = true)
		   AND NOT (sender_id <> $2 AND deleted_by_receiver = true)`,
		conversationID, readerID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, conversation_id, sender_id, content, type, read_at, reply_to_id, created_at
		 FROM messages
		 WHERE conversation_id = $1
		   AND NOT (sender_id = $2 AND deleted_by_sender = true)
		   AND NOT (sender_id <> $2 AND deleted_by_receiver = true)
		 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		conversationID, readerID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*model.Message
	var msgIDs []string
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Type, &m.ReadAt, &m.ReplyToID, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		messages = append(messages, &m)
		msgIDs = append(msgIDs, m.ID)
	}
	if messages == nil {
		messages = []*model.Message{}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Load reply_to for messages that have reply_to_id
	for _, m := range messages {
		r.loadReplyTo(ctx, m)
	}

	// Load reactions for all messages in batch
	if len(msgIDs) > 0 {
		r.loadReactions(ctx, messages, msgIDs)
	}

	return messages, total, nil
}

func (r *ConversationRepo) loadReactions(ctx context.Context, messages []*model.Message, msgIDs []string) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, message_id, user_id, emoji, created_at
		 FROM message_reactions WHERE message_id = ANY($1::uuid[])
		 ORDER BY created_at`, msgIDs,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	reactionMap := make(map[string][]model.MessageReaction)
	for rows.Next() {
		var r model.MessageReaction
		if err := rows.Scan(&r.ID, &r.MessageID, &r.UserID, &r.Emoji, &r.CreatedAt); err != nil {
			continue
		}
		reactionMap[r.MessageID] = append(reactionMap[r.MessageID], r)
	}
	for _, m := range messages {
		if reactions, ok := reactionMap[m.ID]; ok {
			m.Reactions = reactions
		}
	}
}

func (r *ConversationRepo) MarkRead(ctx context.Context, conversationID, readerID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET read_at = NOW()
		 WHERE conversation_id = $1 AND sender_id <> $2 AND read_at IS NULL`,
		conversationID, readerID,
	)
	return err
}

func (r *ConversationRepo) DeleteMessage(ctx context.Context, messageID string, asSender bool) error {
	if asSender {
		_, err := r.pool.Exec(ctx,
			`UPDATE messages SET deleted_by_sender = true WHERE id = $1`, messageID,
		)
		return err
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET deleted_by_receiver = true WHERE id = $1`, messageID,
	)
	return err
}

func (r *ConversationRepo) DeleteMessages(ctx context.Context, messageIDs []string, asSender bool) error {
	if asSender {
		_, err := r.pool.Exec(ctx,
			`UPDATE messages SET deleted_by_sender = true WHERE id = ANY($1::uuid[])`, messageIDs,
		)
		return err
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET deleted_by_receiver = true WHERE id = ANY($1::uuid[])`, messageIDs,
	)
	return err
}

func (r *ConversationRepo) RecallMessages(ctx context.Context, messageIDs []string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET content = 'Tin nhắn đã được thu hồi', type = 'recalled'
		 WHERE id = ANY($1::uuid[]) AND created_at > NOW() - INTERVAL '24 hours'`,
		messageIDs,
	)
	return err
}

func (r *ConversationRepo) RecallMessage(ctx context.Context, messageID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET content = 'Tin nhắn đã được thu hồi', type = 'recalled' WHERE id = $1`,
		messageID,
	)
	return err
}

func (r *ConversationRepo) GetMessageByID(ctx context.Context, messageID string) (*model.Message, error) {
	var m model.Message
	err := r.pool.QueryRow(ctx,
		`SELECT id, conversation_id, sender_id, content, type, read_at, reply_to_id, created_at
		 FROM messages WHERE id = $1`, messageID,
	).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Type, &m.ReadAt, &m.ReplyToID, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("message not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *ConversationRepo) IsParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM conversations
			WHERE id = $1 AND (member_id = $2 OR seller_id = $2)
		)`, conversationID, userID,
	).Scan(&exists)
	return exists, err
}

// --- Reactions ---

func (r *ConversationRepo) ToggleReaction(ctx context.Context, messageID, userID, emoji string) (bool, error) {
	// Try to delete first (toggle off)
	result, err := r.pool.Exec(ctx,
		`DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`,
		messageID, userID, emoji,
	)
	if err != nil {
		return false, err
	}
	if result.RowsAffected() > 0 {
		return false, nil // removed
	}

	// Insert (toggle on)
	_, err = r.pool.Exec(ctx,
		`INSERT INTO message_reactions (message_id, user_id, emoji) VALUES ($1, $2, $3)
		 ON CONFLICT (message_id, user_id, emoji) DO NOTHING`,
		messageID, userID, emoji,
	)
	if err != nil {
		return false, err
	}
	return true, nil // added
}

func (r *ConversationRepo) GetReactionsByMessage(ctx context.Context, messageID string) ([]model.MessageReaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, message_id, user_id, emoji, created_at
		 FROM message_reactions WHERE message_id = $1 ORDER BY created_at`, messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []model.MessageReaction
	for rows.Next() {
		var r model.MessageReaction
		if err := rows.Scan(&r.ID, &r.MessageID, &r.UserID, &r.Emoji, &r.CreatedAt); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	if reactions == nil {
		reactions = []model.MessageReaction{}
	}
	return reactions, rows.Err()
}
