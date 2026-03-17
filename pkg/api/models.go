package api

// JsonRpcRequest represents a standard JSON-RPC 2.0 request
type JsonRpcRequest struct {
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	JSONRPC string      `json:"jsonrpc"`
}

// JsonRpcResponse represents a standard JSON-RPC 2.0 response
type JsonRpcResponse struct {
	ID      string        `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
	JSONRPC string        `json:"jsonrpc"`
}

// JsonRpcError represents a JSON-RPC 2.0 error
type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AuthParams represents parameters for the authenticate method
type AuthParams struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Client   string `json:"client"`
}

// AuthResult represents the result of the authenticate method
type AuthResult struct {
	SessionID  string `json:"sessionId"`
	PersonType int    `json:"personType"`
	PersonID   int    `json:"personId"`
	KlasseID   int    `json:"klasseId"`
}

// BaseEntity represents common fields for WebUntis entities
type BaseEntity struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LongName string `json:"longName"`
	Active   bool   `json:"active,omitempty"`
}

// Klasse represents a class (Klasse)
type Klasse struct {
	BaseEntity
	Did int `json:"did,omitempty"`
}

// Teacher represents a teacher
type Teacher struct {
	BaseEntity
	Title     string `json:"title,omitempty"`
	FirstName string `json:"foreName,omitempty"` // API uses 'foreName' sometimes
	LastName  string `json:"surName,omitempty"`  // API uses 'surName' sometimes
}

// Subject represents a subject
type Subject struct {
	BaseEntity
	AlternateName string `json:"alternateName,omitempty"`
}

// Room represents a room
type Room struct {
	BaseEntity
	Building string `json:"building,omitempty"`
}

// TimetableEntry represents a single lesson/period
type TimetableEntry struct {
	ID           int          `json:"id"`
	Date         int          `json:"date"`      // YYYYMMDD
	StartTime    int          `json:"startTime"` // HHMM
	EndTime      int          `json:"endTime"`   // HHMM
	Kl           []BaseEntity `json:"kl"`
	Te           []BaseEntity `json:"te"`
	Su           []BaseEntity `json:"su"`
	Ro           []BaseEntity `json:"ro"`
	ActivityType string       `json:"activityType,omitempty"`
	Code         string       `json:"code,omitempty"` // "cancelled", "irregular", etc.
	SubstText    string       `json:"substText,omitempty"`
	Statflags    string       `json:"statflags,omitempty"`
}

// TimetableParams represents parameters for getTimetable
type TimetableParams struct {
	Options TimetableOptions `json:"options"`
}

type TimetableOptions struct {
	Element   TimetableElement `json:"element"`
	StartDate int              `json:"startDate"` // YYYYMMDD
	EndDate   int              `json:"endDate"`   // YYYYMMDD
	ShowInfo  bool             `json:"showInfo,omitempty"`
}

type TimetableElement struct {
	ID   int `json:"id"`
	Type int `json:"type"` // 1=Class, 2=Teacher, 3=Subject, 4=Room, 5=Student
}

// StudentInfo represents a student accessible to the logged-in user
type StudentInfo struct {
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
}

// AppDataResponse is the structure for the /api/rest/view/v1/app/data endpoint
type AppDataResponse struct {
	Tenant struct {
		ID string `json:"id"`
	} `json:"tenant"`
	CurrentSchoolYear struct {
		ID int `json:"id"`
	} `json:"currentSchoolYear"`
	User struct {
		Students []StudentInfo `json:"students"`
	} `json:"user"`
}

// --- REST Timetable Models (/api/rest/view/v1/timetable/entries) ---

// RestTimetableResponse is the top-level response from the REST timetable endpoint
type RestTimetableResponse struct {
	Days []RestDay `json:"days"`
}

// RestDay represents a single day in the timetable response
type RestDay struct {
	Date        string          `json:"date"`
	Status      string          `json:"status"`
	GridEntries []RestGridEntry `json:"gridEntries"`
}

// RestGridEntry represents a single lesson/period entry
type RestGridEntry struct {
	IDs              []int          `json:"ids"`
	Duration         RestDuration   `json:"duration"`
	Type             string         `json:"type"`
	Status           string         `json:"status"`
	Position1        []RestPosition `json:"position1"` // Teacher
	Position2        []RestPosition `json:"position2"` // Subject
	Position3        []RestPosition `json:"position3"` // Room
	Color            string         `json:"color"`
	SubstitutionText string         `json:"substitutionText"`
	LessonText       string         `json:"lessonText"`
	NotesAll         string         `json:"notesAll"`
}

// RestDuration holds the start and end ISO datetime strings
type RestDuration struct {
	Start string `json:"start"` // e.g. "2026-03-09T08:30"
	End   string `json:"end"`
}

// RestPosition holds current and optionally removed entity info (for substitutions)
type RestPosition struct {
	Current *RestPositionEntry `json:"current"`
	Removed *RestPositionEntry `json:"removed"`
}

// RestPositionEntry is a single entity (teacher, subject, room) in a position slot
type RestPositionEntry struct {
	Type        string `json:"type"` // TEACHER, SUBJECT, ROOM
	Status      string `json:"status"`
	ShortName   string `json:"shortName"`
	LongName    string `json:"longName"`
	DisplayName string `json:"displayName"`
}

// --- Absences Models (/api/classreg/absences/students) ---

// AbsencesResponse is the top-level response from the absences endpoint
type AbsencesResponse struct {
	Data struct {
		Absences []Absence `json:"absences"`
	} `json:"data"`
}

// Absence represents a single absence record
type Absence struct {
	ID           int    `json:"id"`
	StartDate    int    `json:"startDate"` // YYYYMMDD
	EndDate      int    `json:"endDate"`   // YYYYMMDD
	StartTime    int    `json:"startTime"` // HHMM
	EndTime      int    `json:"endTime"`   // HHMM
	ReasonID     int    `json:"reasonId"`
	Reason       string `json:"reason"`
	Text         string `json:"text"`
	StudentName  string `json:"studentName"`
	ExcuseStatus string `json:"excuseStatus"`
	IsExcused    bool   `json:"isExcused"`
	CanEdit      bool   `json:"canEdit"`
}

// FlatAbsence is the CLI output format for absences
type FlatAbsence struct {
	Date         string `json:"date"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	Reason       string `json:"reason"`
	Text         string `json:"text,omitempty"`
	StudentName  string `json:"studentName"`
	ExcuseStatus string `json:"excuseStatus"`
	IsExcused    bool   `json:"isExcused"`
}

// --- Homework Models (/api/homeworks/lessons) ---

// HomeworkResponse is the top-level response from the homework endpoint
type HomeworkResponse struct {
	Data struct {
		Records   []HomeworkRecord  `json:"records"`
		Homeworks []Homework        `json:"homeworks"`
		Teachers  []HomeworkTeacher `json:"teachers"`
		Lessons   []HomeworkLesson  `json:"lessons"`
	} `json:"data"`
}

// HomeworkRecord maps a homework assignment to a teacher and student(s)
type HomeworkRecord struct {
	HomeworkID int   `json:"homeworkId"`
	TeacherID  int   `json:"teacherId"`
	ElementIDs []int `json:"elementIds"`
}

// Homework represents a single homework assignment
type Homework struct {
	ID        int    `json:"id"`
	LessonID  int    `json:"lessonId"`
	Date      int    `json:"date"`    // YYYYMMDD
	DueDate   int    `json:"dueDate"` // YYYYMMDD
	Text      string `json:"text"`
	Remark    string `json:"remark"`
	Completed bool   `json:"completed"`
}

// HomeworkTeacher maps teacher ID to name in the homework response
type HomeworkTeacher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// HomeworkLesson maps lesson ID to subject in the homework response
type HomeworkLesson struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
}

// FlatHomework is the CLI output format for homework
type FlatHomework struct {
	Date      string `json:"date"`
	DueDate   string `json:"dueDate"`
	Text      string `json:"text"`
	Teacher   string `json:"teacher,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Remark    string `json:"remark,omitempty"`
	Completed bool   `json:"completed"`
}

// --- Messages Models (/api/rest/view/v1/messages) ---

// MessagesResponse is the top-level response from the messages endpoint
type MessagesResponse struct {
	IncomingMessages []Message `json:"incomingMessages"`
}

// Message represents a single inbox message
type Message struct {
	ID                   int           `json:"id"`
	Subject              string        `json:"subject"`
	ContentPreview       string        `json:"contentPreview"`
	Sender               MessageSender `json:"sender"`
	SentDateTime         string        `json:"sentDateTime"`
	IsMessageRead        bool          `json:"isMessageRead"`
	IsReply              bool          `json:"isReply"`
	IsReplyAllowed       bool          `json:"isReplyAllowed"`
	HasAttachments       bool          `json:"hasAttachments"`
	AllowMessageDeletion bool          `json:"allowMessageDeletion"`
}

// MessageSender holds sender info for a message
type MessageSender struct {
	DisplayName string `json:"displayName"`
	UserID      int    `json:"userId"`
}

// FlatMessage is the CLI output format for messages
type FlatMessage struct {
	Subject        string `json:"subject"`
	Sender         string `json:"sender"`
	SentAt         string `json:"sentAt"`
	ContentPreview string `json:"contentPreview"`
	IsRead         bool   `json:"isRead"`
	HasAttachments bool   `json:"hasAttachments,omitempty"`
}

// FlatTimetableEntry is the flat CLI output format for timetable entries
type FlatTimetableEntry struct {
	Date             string `json:"date"`
	Start            string `json:"start"`
	End              string `json:"end"`
	Subject          string `json:"subject"`
	SubjectShort     string `json:"subjectShort"`
	Teacher          string `json:"teacher"`
	TeacherShort     string `json:"teacherShort"`
	Room             string `json:"room"`
	Status           string `json:"status"`
	SubstitutionText string `json:"substitutionText,omitempty"`
	LessonText       string `json:"lessonText,omitempty"`
	Color            string `json:"color,omitempty"`
}
