// MongoDB initialization script
// Runs automatically when container is created for the first time.

// Switch to rice_chat database
db = db.getSiblingDB('rice_chat');

// Create collections with schema validation
db.createCollection('messages', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['conversation_id', 'sender_id', 'content', 'type', 'timestamp'],
      properties: {
        conversation_id: { bsonType: 'string', description: 'ID cuộc hội thoại' },
        sender_id: { bsonType: 'string', description: 'UUID người gửi' },
        content: { bsonType: 'string', description: 'Nội dung tin nhắn hoặc URL ảnh' },
        type: { enum: ['text', 'image'], description: 'Loại tin nhắn' },
        timestamp: { bsonType: 'date', description: 'Thời gian gửi' },
        read_at: { bsonType: ['date', 'null'], description: 'Thời gian đã đọc' }
      }
    }
  }
});

db.createCollection('conversations', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['participants', 'created_at'],
      properties: {
        participants: {
          bsonType: 'array',
          minItems: 2,
          maxItems: 2,
          items: { bsonType: 'string' },
          description: 'UUID của 2 người tham gia'
        },
        listing_id: { bsonType: ['string', 'null'], description: 'Tin đăng liên quan' },
        last_message_at: { bsonType: ['date', 'null'] },
        created_at: { bsonType: 'date' }
      }
    }
  }
});

// Create indexes
db.messages.createIndex({ conversation_id: 1, timestamp: -1 });
db.messages.createIndex({ sender_id: 1 });
db.messages.createIndex({ conversation_id: 1, sender_id: 1, read_at: 1 });

db.conversations.createIndex({ participants: 1 });
db.conversations.createIndex({ last_message_at: -1 });

print('✅ Rice Chat MongoDB initialized successfully!');
