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
	err := r.pool.QueryRow(ctx,
		`SELECT id, buyer_id, seller_id, listing_id, last_message_at, created_at
		 FROM conversations
		 WHERE (buyer_id = $1 AND seller_id = $2) OR (buyer_id = $2 AND seller_id = $1)
		 ORDER BY last_message_at DESC LIMIT 1`,
		buyerID, sellerID,
	).Scan(&conv.ID, &conv.BuyerID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
	if err == nil {
		return &conv, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Create new
	err = r.pool.QueryRow(ctx,
		`INSERT INTO conversations (buyer_id, seller_id, listing_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, buyer_id, seller_id, listing_id, last_message_at, created_at`,
		buyerID, sellerID, listingID,
	).Scan(&conv.ID, &conv.BuyerID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) GetByID(ctx context.Context, id string) (*model.Conversation, error) {
	var conv model.Conversation
	err := r.pool.QueryRow(ctx,
		`SELECT id, buyer_id, seller_id, listing_id, last_message_at, created_at
		 FROM conversations WHERE id = $1`,
		id,
	).Scan(&conv.ID, &conv.BuyerID, &conv.SellerID, &conv.ListingID, &conv.LastMessageAt, &conv.CreatedAt)
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
		`SELECT COUNT(*) FROM conversations WHERE buyer_id = $1 OR seller_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT c.id, c.buyer_id, c.seller_id, c.listing_id, c.last_message_at, c.created_at,
		        u.id, u.role, u.name, u.avatar_url, u.province, u.description, u.org_name, u.created_at,
		        COALESCE(unread.cnt, 0)
		 FROM conversations c
		 JOIN users u ON u.id = CASE WHEN c.buyer_id = $1 THEN c.seller_id ELSE c.buyer_id END
		 LEFT JOIN (
		     SELECT conversation_id, COUNT(*) AS cnt
		     FROM messages
		     WHERE sender_id <> $1 AND read_at IS NULL
		     GROUP BY conversation_id
		 ) unread ON unread.conversation_id = c.id
		 WHERE c.buyer_id = $1 OR c.seller_id = $1
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
			&conv.ID, &conv.BuyerID, &conv.SellerID, &conv.ListingID,
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

func (r *ConversationRepo) SendMessage(ctx context.Context, conversationID, senderID, content, msgType string) (*model.Message, error) {
	if msgType == "" {
		msgType = "text"
	}

	var msg model.Message
	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, sender_id, content, type)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, conversation_id, sender_id, content, type, read_at, created_at`,
		conversationID, senderID, content, msgType,
	).Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Content, &msg.Type, &msg.ReadAt, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Update last_message_at
	_, _ = r.pool.Exec(ctx,
		`UPDATE conversations SET last_message_at = NOW() WHERE id = $1`, conversationID,
	)

	return &msg, nil
}

func (r *ConversationRepo) GetMessages(ctx context.Context, conversationID, readerID string, page, limit int) ([]*model.Message, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages
		 WHERE conversation_id = $1
		   AND NOT (sender_id = $2 AND deleted_by_sender = true)`,
		conversationID, readerID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, conversation_id, sender_id, content, type, read_at, created_at
		 FROM messages
		 WHERE conversation_id = $1
		   AND NOT (sender_id = $2 AND deleted_by_sender = true)
		 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
		conversationID, readerID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Type, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		messages = append(messages, &m)
	}
	if messages == nil {
		messages = []*model.Message{}
	}
	return messages, total, rows.Err()
}

func (r *ConversationRepo) MarkRead(ctx context.Context, conversationID, readerID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET read_at = NOW()
		 WHERE conversation_id = $1 AND sender_id <> $2 AND read_at IS NULL`,
		conversationID, readerID,
	)
	return err
}

func (r *ConversationRepo) DeleteMessage(ctx context.Context, messageID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET deleted_by_sender = true WHERE id = $1`, messageID,
	)
	return err
}

func (r *ConversationRepo) DeleteMessages(ctx context.Context, messageIDs []string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE messages SET deleted_by_sender = true WHERE id = ANY($1::uuid[])`, messageIDs,
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
		`SELECT id, conversation_id, sender_id, content, type, read_at, created_at
		 FROM messages WHERE id = $1`, messageID,
	).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Type, &m.ReadAt, &m.CreatedAt)
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
			WHERE id = $1 AND (buyer_id = $2 OR seller_id = $2)
		)`, conversationID, userID,
	).Scan(&exists)
	return exists, err
}
