package api

// RefSource describes how to obtain candidate values for an ID field.
//
// This is the core idea behind gh's prompts: never ask the user to invent a
// value — fetch real data first and let them choose. gh pr create lists
// repository collaborators when asking for a reviewer; coj match rank lists
// real matches when asking for F_MatchID.
type RefSource struct {
	Field     string   // target field, e.g. F_MatchID
	ListPath  string   // endpoint used to fetch candidates
	LabelKeys []string // candidate label keys, in priority order
	IDKey     string   // key holding the value
	// Extra holds fixed fields some endpoints require when fetching
	Extra map[string]string
}

// RefSources lists the known candidate sources for ID fields.
var RefSources = []RefSource{
	{Field: "F_MatchID", ListPath: "/Match/getMatchList", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_ClassID", ListPath: "/Fclass/getFclassList", LabelKeys: []string{"F_Name"}, IDKey: "F_ID"},
	{Field: "F_SubjectID", ListPath: "/Subject/getSubjectListNew", LabelKeys: []string{"F_Title", "F_Name"}, IDKey: "F_ID"},
	{Field: "F_CourseID", ListPath: "/Course/getCourseList", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_HomeworkID", ListPath: "/Homework/getHomeworkList", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_ExamID", ListPath: "/Exam/getExamListManager", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_ChapterID", ListPath: "/Chapter/getChapterList", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_UserID", ListPath: "/User/getUserList", LabelKeys: []string{"F_Name", "F_UserName", "F_Account"}, IDKey: "F_ID"},
	{Field: "F_BagID", ListPath: "/Bag/getBagList", LabelKeys: []string{"F_Name", "F_Title"}, IDKey: "F_ID"},
	{Field: "F_TagID", ListPath: "/Tag/getTagList", LabelKeys: []string{"F_Name"}, IDKey: "F_ID"},
	{Field: "F_OriginID", ListPath: "/Origin/getOriginList", LabelKeys: []string{"F_Name"}, IDKey: "F_ID"},
	{Field: "F_DifficultyID", ListPath: "/Difficulty/getDifficultyList", LabelKeys: []string{"F_Name"}, IDKey: "F_ID"},
	{Field: "F_MasterID", ListPath: "/User/getMasterList", LabelKeys: []string{"F_Name", "F_UserName"}, IDKey: "F_ID"},
	{Field: "F_SchoolUserID", ListPath: "/User/getMasterList", LabelKeys: []string{"F_Name", "F_UserName"}, IDKey: "F_ID"},
}

// refIndex speeds up lookups.
var refIndex = func() map[string]RefSource {
	m := make(map[string]RefSource, len(RefSources))
	for _, r := range RefSources {
		m[r.Field] = r
	}
	return m
}()

// RefFor returns the candidate source for a field.
func RefFor(field string) (RefSource, bool) {
	r, ok := refIndex[field]
	return r, ok
}

// IsIDField reports whether a field name looks like an ID reference.
func IsIDField(field string) bool {
	if len(field) < 3 {
		return false
	}
	return field[len(field)-2:] == "ID"
}
