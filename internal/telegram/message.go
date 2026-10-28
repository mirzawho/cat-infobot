package telegram

import "encoding/json"

// Message is the Bot API message. Fields the bot does not display are kept in
// Raw so that nothing Telegram sent is lost; unknown future fields therefore
// never break parsing.
type Message struct {
	MessageID      int             `json:"message_id"`
	Date           int64           `json:"date"`
	EditDate       int64           `json:"edit_date"`
	From           *User           `json:"from"`
	Chat           Chat            `json:"chat"`
	Text           string          `json:"text"`
	Caption        string          `json:"caption"`
	Entities       []Entity        `json:"entities"`
	CaptionEntity  []Entity        `json:"caption_entities"`
	ForwardOrigin  *ForwardOrigin  `json:"forward_origin"`
	ForwardFrom    *User           `json:"forward_from"`
	ForwardFromCh  *Chat           `json:"forward_from_chat"`
	ForwardDate    int64           `json:"forward_date"`
	ForwardSig     string          `json:"forward_signature"`
	ForwardName    string          `json:"forward_sender_name"`
	ReplyTo        *Message        `json:"reply_to_message"`
	IsTopicMessage bool            `json:"is_topic_message"`
	MessageThread  int             `json:"message_thread_id"`
	MediaGroupID   string          `json:"media_group_id"`
	AuthorSig      string          `json:"author_signature"`
	ViaBot         *User           `json:"via_bot"`
	Photo          []PhotoSize     `json:"photo"`
	Video          *Video          `json:"video"`
	Document       *Document       `json:"document"`
	Audio          *Audio          `json:"audio"`
	Voice          *Document       `json:"voice"`
	VideoNote      *Document       `json:"video_note"`
	Animation      *Document       `json:"animation"`
	Sticker        *Sticker        `json:"sticker"`
	Contact        json.RawMessage `json:"contact"`
	Location       json.RawMessage `json:"location"`
	Venue          json.RawMessage `json:"venue"`
	Poll           json.RawMessage `json:"poll"`
	Dice           json.RawMessage `json:"dice"`
	Invoice        json.RawMessage `json:"invoice"`
	Raw            json.RawMessage `json:"-"`
}

// UnmarshalJSON decodes the known fields and keeps a copy of the entire
// object in Raw.
func (m *Message) UnmarshalJSON(data []byte) error {
	type plain Message
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	*m = Message(p)
	m.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// User is a Telegram account.
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// Chat identifies a conversation.
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Entity marks a tagged range of text (bold, link, mention, ...).
type Entity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
}

// ForwardOrigin is where a forwarded message came from.
type ForwardOrigin struct {
	Type            string `json:"type"`
	Date            int64  `json:"date"`
	Chat            *Chat  `json:"chat"`
	SenderUser      *User  `json:"sender_user"`
	SenderUserName  string `json:"sender_user_name"`
	MessageID       int    `json:"message_id"`
	AuthorSignature string `json:"author_signature"`
}

// PhotoSize is one size of a photo.
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int64  `json:"file_size"`
}

// Video is a video message or animation.
type Video struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Duration     int    `json:"duration"`
	MimeType     string `json:"mime_type"`
	FileSize     int64  `json:"file_size"`
	FileName     string `json:"file_name"`
}

// Audio is an audio message.
type Audio struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Duration     int    `json:"duration"`
	Performer    string `json:"performer"`
	Title        string `json:"title"`
	FileName     string `json:"file_name"`
	MimeType     string `json:"mime_type"`
	FileSize     int64  `json:"file_size"`
}

// Document is a file, and also the payload type of voice and video notes.
type Document struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name"`
	MimeType     string `json:"mime_type"`
	FileSize     int64  `json:"file_size"`
	Duration     int    `json:"duration"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

// Sticker is a sticker.
type Sticker struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	IsAnimated   bool   `json:"is_animated"`
	IsVideo      bool   `json:"is_video"`
	Emoji        string `json:"emoji"`
	SetName      string `json:"set_name"`
	FileSize     int64  `json:"file_size"`
}
