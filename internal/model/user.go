package model

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Username  string    `json:"username" gorm:"size:50;not null;uniqueIndex:idx_users_username;index:idx_users_username;column:username"`
	Password  string    `json:"password" gorm:"size:255;not null;column:password"`
	Email     string    `json:"email" gorm:"size:255;not null;uniqueIndex:idx_users_email;index:idx_users_email;column:email"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type Chat struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `json:"name,omitempty" gorm:"size:255;column:name"`
	IsGroup   bool      `json:"is_group" gorm:"column:is_group;default:false"`
	CreatedBy int       `json:"created_by" gorm:"not null;column:created_by;index:idx_chats_created_by"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	Creator   *User     `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnDelete:CASCADE"`
}

type ChatParticipant struct {
	ID       int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ChatID   int       `json:"chat_id" gorm:"not null;column:chat_id;index:idx_chats_participants_chat_id;uniqueIndex:idx_chat_user"`
	UserID   int       `json:"user_id" gorm:"not null;column:user_id;index:idx_chats_participants_user_id;uniqueIndex:idx_chat_user"`
	JoinedAt time.Time `json:"joined_at" gorm:"column:joined_at;autoCreateTime"`
	Chat     *Chat     `json:"chat,omitempty" gorm:"foreignKey:ChatID;references:ID;constraint:OnDelete:CASCADE"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (ChatParticipant) TableName() string {
	return "chats_participants"
}

type Message struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ChatID    int       `json:"chat_id" gorm:"not null;column:chat_id;index:idx_messages_chat_id"`
	SenderID  int       `json:"sender_id" gorm:"not null;column:sender_id;index:idx_messages_sender_id"`
	Content   string    `json:"content" gorm:"type:text;not null;column:content"`
	IsRead    bool      `json:"is_read" gorm:"column:is_read;default:false;index:idx_messages_is_read"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime;index:idx_messages_created_at"`
	Chat      *Chat     `json:"chat,omitempty" gorm:"foreignKey:ChatID;references:ID;constraint:OnDelete:CASCADE"`
	Sender    *User     `json:"sender,omitempty" gorm:"foreignKey:SenderID;references:ID;constraint:OnDelete:CASCADE"`
}
