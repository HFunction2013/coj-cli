// Code generated from CandyOJ production bundle. DO NOT EDIT.

package api

// WebTarget 描述一个接口对应的网页入口，供 `coj ... --web` 打开浏览器使用。
//
// 与 gh 的 --web 同一思路：命令行负责筛选与定位，
// 需要富交互或富展示的操作交给网页完成。
type WebTarget struct {
	Path     string // 网页路径，可能含 :id 占位符
	IDField  string // 用于替换 :id 的请求字段名，为空表示无需 id
	Resource string
	Action   string
}

// WebTargets 接口路径 -> 网页入口
var WebTargets = map[string]WebTarget{
	"/Bag/addBagForMaster":                     {Path: "/TestPaperEdit/:id", IDField: "F_BagID", Resource: "bag", Action: "add-bag-for-master"},
	"/Bag/addBagType":                          {Path: "/TestPaperEdit/:id", IDField: "F_ID", Resource: "bag", Action: "add-bag-type"},
	"/Bag/bagDel":                              {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "bag-del"},
	"/Bag/bagPartSave":                         {Path: "/TestPaperEdit/:id", IDField: "", Resource: "bag", Action: "bag-part-save"},
	"/Bag/bagSave":                             {Path: "/TestPaperEdit/:id", IDField: "", Resource: "bag", Action: "bag-save"},
	"/Bag/bagStatus":                           {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "bag-status"},
	"/Bag/delBagType":                          {Path: "/TestPaperEdit/:id", IDField: "", Resource: "bag", Action: "del-bag-type"},
	"/Bag/getBagDetail":                        {Path: "/TestPaperDesc/:id", IDField: "F_ID", Resource: "bag", Action: "get-bag-detail"},
	"/Bag/getBagList":                          {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "get-bag-list"},
	"/Bag/getBagTypeList":                      {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "get-bag-type-list"},
	"/Bag/getBagTypeListForManager":            {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "get-bag-type-list-for-manager"},
	"/Bag/getClassMatchList":                   {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "get-class-match-list"},
	"/Bag/getMasterBag":                        {Path: "/TestPaper", IDField: "", Resource: "bag", Action: "get-master-bag"},
	"/Chapter/addExamToChapter":                {Path: "/ChapterDesc/:id", IDField: "F_ChapterID", Resource: "chapter", Action: "add-exam-to-chapter"},
	"/Chapter/chapterExamRank":                 {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "chapter-exam-rank"},
	"/Chapter/chapterPartSave":                 {Path: "/ChapterDesc/:id", IDField: "", Resource: "chapter", Action: "chapter-part-save"},
	"/Chapter/chapterRank":                     {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "chapter-rank"},
	"/Chapter/chapterSave":                     {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "chapter-save"},
	"/Chapter/delExamFromChapter":              {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "del-exam-from-chapter"},
	"/Chapter/deleteChapter":                   {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "delete-chapter"},
	"/Chapter/getChapterDetail":                {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter", Action: "get-chapter-detail"},
	"/Chapter/getChapterList":                  {Path: "/ChapterList/:id", IDField: "F_CourseID", Resource: "chapter", Action: "get-chapter-list"},
	"/Chapter/uploadChapterFile":               {Path: "/ChapterDesc/:id", IDField: "", Resource: "chapter", Action: "upload-chapter-file"},
	"/ChapterSubject/addChapterSubjectType":    {Path: "/ChapterDesc/:id", IDField: "", Resource: "chapter-subject", Action: "add-chapter-subject-type"},
	"/ChapterSubject/addSubjectToChapter":      {Path: "/ChapterDesc/:id", IDField: "F_ChapterID", Resource: "chapter-subject", Action: "add-subject-to-chapter"},
	"/ChapterSubject/addSubjectToChapterNew":   {Path: "/ChapterDesc/:id", IDField: "F_ChapterID", Resource: "chapter-subject", Action: "add-subject-to-chapter-new"},
	"/ChapterSubject/delChapterSubjectType":    {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter-subject", Action: "del-chapter-subject-type"},
	"/ChapterSubject/deletechaptersubject":     {Path: "/ChapterDesc/:id", IDField: "F_ID", Resource: "chapter-subject", Action: "deletechaptersubject"},
	"/ChapterSubject/getAllSubject":            {Path: "/ChapterDesc/:id", IDField: "", Resource: "chapter-subject", Action: "get-all-subject"},
	"/ChapterSubject/getChapterSubjectListNew": {Path: "/ChapterList/:id", IDField: "F_ChapterID", Resource: "chapter-subject", Action: "get-chapter-subject-list-new"},
	"/ChapterSubject/getSubjectForAdd":         {Path: "/ChapterDesc/:id", IDField: "F_ChapterID", Resource: "chapter-subject", Action: "get-subject-for-add"},
	"/Course/addCourseForMaster":               {Path: "/CourseManage", IDField: "", Resource: "course", Action: "add-course-for-master"},
	"/Course/courseExamRank":                   {Path: "/CourseManage", IDField: "", Resource: "course", Action: "course-exam-rank"},
	"/Course/courseRank":                       {Path: "/CourseManage", IDField: "", Resource: "course", Action: "course-rank"},
	"/Course/courseSave":                       {Path: "/CourseManage", IDField: "", Resource: "course", Action: "course-save"},
	"/Course/deleteCourse":                     {Path: "/CourseManage", IDField: "", Resource: "course", Action: "delete-course"},
	"/Course/getCanAddCourse":                  {Path: "/CourseManage", IDField: "", Resource: "course", Action: "get-can-add-course"},
	"/Course/getCourseChapter":                 {Path: "/CourseManage", IDField: "", Resource: "course", Action: "get-course-chapter"},
	"/Course/getCourseList":                    {Path: "/CourseManage", IDField: "", Resource: "course", Action: "get-course-list"},
	"/Course/getCourseStudent":                 {Path: "/CourseManage", IDField: "", Resource: "course", Action: "get-course-student"},
	"/Course/getMasterCourse":                  {Path: "/CourseManage", IDField: "", Resource: "course", Action: "get-master-course"},
	"/Course/getStudentClass":                  {Path: "/MyCourse", IDField: "", Resource: "course", Action: "get-student-class"},
	"/Course/setClassCourse":                   {Path: "/CourseManage", IDField: "", Resource: "course", Action: "set-class-course"},
	"/Difficulty/deleteDifficulty":             {Path: "/TagsDifficulty", IDField: "", Resource: "difficulty", Action: "delete-difficulty"},
	"/Difficulty/getDifficultyList":            {Path: "/TagsDifficulty", IDField: "", Resource: "difficulty", Action: "get-difficulty-list"},
	"/Difficulty/saveDifficulty":               {Path: "/TagsDifficulty", IDField: "", Resource: "difficulty", Action: "save-difficulty"},
	"/Exam/addExamSubject":                     {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "add-exam-subject"},
	"/Exam/deleteExam":                         {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "delete-exam"},
	"/Exam/deleteExamSubject":                  {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "delete-exam-subject"},
	"/Exam/getExamList":                        {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-list"},
	"/Exam/getExamListManager":                 {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-list-manager"},
	"/Exam/getExamLog":                         {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-log"},
	"/Exam/getExamLogListNew":                  {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-log-list-new"},
	"/Exam/getExamLogSubject":                  {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-log-subject"},
	"/Exam/getExamSubject":                     {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-exam-subject"},
	"/Exam/getStudentExamLogList":              {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "get-student-exam-log-list"},
	"/Exam/saveExam":                           {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "save-exam"},
	"/Exam/saveImg":                            {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "save-img"},
	"/Exam/takeExam":                           {Path: "/QuestionBank", IDField: "", Resource: "exam", Action: "take-exam"},
	"/Fclass/addStudenToClass":                 {Path: "/ClassManage", IDField: "", Resource: "class", Action: "add-studen-to-class"},
	"/Fclass/addTeacherFromClass":              {Path: "/ClassManage", IDField: "", Resource: "class", Action: "add-teacher-from-class"},
	"/Fclass/deleteClass":                      {Path: "/ClassManage", IDField: "", Resource: "class", Action: "delete-class"},
	"/Fclass/getCanMoveClassList":              {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-can-move-class-list"},
	"/Fclass/getClassMateRanking":              {Path: "/MyClassOrder/:id", IDField: "F_ClassID", Resource: "class", Action: "get-class-mate-ranking"},
	"/Fclass/getClassStudentList":              {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-class-student-list"},
	"/Fclass/getFclassList":                    {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-fclass-list"},
	"/Fclass/getMasterClass":                   {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-master-class"},
	"/Fclass/getRanking":                       {Path: "/MyClassOrder/:id", IDField: "F_ClassID", Resource: "class", Action: "get-ranking"},
	"/Fclass/getStudentClassList":              {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-student-class-list"},
	"/Fclass/getTeacherCanMoveClass":           {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-teacher-can-move-class"},
	"/Fclass/getTeacherClass":                  {Path: "/ClassManage", IDField: "", Resource: "class", Action: "get-teacher-class"},
	"/Fclass/removeClassByTeacher":             {Path: "/ClassManage", IDField: "", Resource: "class", Action: "remove-class-by-teacher"},
	"/Fclass/removeStudentFromClass":           {Path: "/ClassManage", IDField: "", Resource: "class", Action: "remove-student-from-class"},
	"/Fclass/saveFclass":                       {Path: "/ClassManage", IDField: "", Resource: "class", Action: "save-fclass"},
	"/Grade/getGradeList":                      {Path: "/ClassManage", IDField: "", Resource: "grade", Action: "get-grade-list"},
	"/Homework/addHomeworkNew":                 {Path: "/HomeworkEdit/:id", IDField: "", Resource: "homework", Action: "add-homework-new"},
	"/Homework/delHomework":                    {Path: "/HomeworkEdit/:id", IDField: "F_ID", Resource: "homework", Action: "del-homework"},
	"/Homework/delPublishScore":                {Path: "/HomeworkDesc/:id", IDField: "F_HomeworkID", Resource: "homework", Action: "del-publish-score"},
	"/Homework/dingzhengHomework":              {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "dingzheng-homework"},
	"/Homework/doHomework":                     {Path: "/HomeworkDesc/:id", IDField: "F_HomeworkID", Resource: "homework", Action: "do-homework"},
	"/Homework/endSelfHomework":                {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "end-self-homework"},
	"/Homework/getHomeworkClass":               {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-homework-class"},
	"/Homework/getHomeworkList":                {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-homework-list"},
	"/Homework/getHomeworkRank":                {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-homework-rank"},
	"/Homework/getHomeworkSubject":             {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-homework-subject"},
	"/Homework/getHomeworkTestLog":             {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-homework-test-log"},
	"/Homework/getSubjectFromHomework":         {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "get-subject-from-homework"},
	"/Homework/hideHomework":                   {Path: "/HomeworkEdit/:id", IDField: "F_ID", Resource: "homework", Action: "hide-homework"},
	"/Homework/homeworkPartSave":               {Path: "/HomeworkEdit/:id", IDField: "", Resource: "homework", Action: "homework-part-save"},
	"/Homework/publishScore":                   {Path: "/HomeworkDesc/:id", IDField: "F_HomeworkID", Resource: "homework", Action: "publish-score"},
	"/Homework/subjectDetail":                  {Path: "/HomeworkDesc/:id", IDField: "F_HomeworkID", Resource: "homework", Action: "subject-detail"},
	"/Homework/submitHomework":                 {Path: "/HomeworkDesc/:id", IDField: "F_HomeworkID", Resource: "homework", Action: "submit-homework"},
	"/Homework/tiQianEnd":                      {Path: "/HomeworkManage", IDField: "", Resource: "homework", Action: "ti-qian-end"},
	"/Match/addMatchNew":                       {Path: "/MatchEdit/:id", IDField: "", Resource: "match", Action: "add-match-new"},
	"/Match/addMatchToClass":                   {Path: "/MatchEdit/:id", IDField: "F_MatchID", Resource: "match", Action: "add-match-to-class"},
	"/Match/addSubjectToMatch":                 {Path: "/MatchEdit/:id", IDField: "F_ID", Resource: "match", Action: "add-subject-to-match"},
	"/Match/canAddList":                        {Path: "/MatchManage", IDField: "", Resource: "match", Action: "can-add-list"},
	"/Match/delMatch":                          {Path: "/MatchEdit/:id", IDField: "F_ID", Resource: "match", Action: "del-match"},
	"/Match/delPublishScore":                   {Path: "/MatchDesc/:id", IDField: "F_MatchID", Resource: "match", Action: "del-publish-score"},
	"/Match/deleteSubjectFromMatch":            {Path: "/MatchEdit/:id", IDField: "F_ID", Resource: "match", Action: "delete-subject-from-match"},
	"/Match/dingzhengMatch":                    {Path: "/MatchManage", IDField: "", Resource: "match", Action: "dingzheng-match"},
	"/Match/doMatch":                           {Path: "/MatchAnswering", IDField: "", Resource: "match", Action: "do-match"},
	"/Match/endSelfMatch":                      {Path: "/MatchManage", IDField: "", Resource: "match", Action: "end-self-match"},
	"/Match/exportoffline":                     {Path: "/MatchEdit/:id", IDField: "F_MatchID", Resource: "match", Action: "exportoffline"},
	"/Match/exportproblemspdf":                 {Path: "/MatchEdit/:id", IDField: "F_MatchID", Resource: "match", Action: "exportproblemspdf"},
	"/Match/exportrecord":                      {Path: "/MatchEdit/:id", IDField: "F_MatchID", Resource: "match", Action: "exportrecord"},
	"/Match/getClassMatch":                     {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-class-match"},
	"/Match/getCodeContent":                    {Path: "/MatchAnswering", IDField: "", Resource: "match", Action: "get-code-content"},
	"/Match/getCondition":                      {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-condition"},
	"/Match/getMatchClass":                     {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-class"},
	"/Match/getMatchList":                      {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-list"},
	"/Match/getMatchRank":                      {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-rank"},
	"/Match/getMatchStatus":                    {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-status"},
	"/Match/getMatchSubject":                   {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-subject"},
	"/Match/getMatchTestLog":                   {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-match-test-log"},
	"/Match/getStudentMatch":                   {Path: "/MyMatchs", IDField: "", Resource: "match", Action: "get-student-match"},
	"/Match/getSubjectFromMatch":               {Path: "/MatchManage", IDField: "", Resource: "match", Action: "get-subject-from-match"},
	"/Match/hideMatch":                         {Path: "/MatchEdit/:id", IDField: "F_ID", Resource: "match", Action: "hide-match"},
	"/Match/importoffline":                     {Path: "/MatchEdit/:id", IDField: "", Resource: "match", Action: "importoffline"},
	"/Match/matchPartSave":                     {Path: "/MatchEdit/:id", IDField: "", Resource: "match", Action: "match-part-save"},
	"/Match/publishScore":                      {Path: "/MatchDesc/:id", IDField: "F_MatchID", Resource: "match", Action: "publish-score"},
	"/Match/showBagExplain":                    {Path: "/MatchManage", IDField: "", Resource: "match", Action: "show-bag-explain"},
	"/Match/studentGetMatchSubject":            {Path: "/MyMatchs", IDField: "", Resource: "match", Action: "student-get-match-subject"},
	"/Match/subjectDetail":                     {Path: "/MatchDesc/:id", IDField: "F_MatchID", Resource: "match", Action: "subject-detail"},
	"/Match/submitMatch":                       {Path: "/MatchAnswering", IDField: "", Resource: "match", Action: "submit-match"},
	"/Match/tiQianEnd":                         {Path: "/MatchManage", IDField: "", Resource: "match", Action: "ti-qian-end"},
	"/Origin/deleteOrigin":                     {Path: "/TagsOrigin", IDField: "", Resource: "origin", Action: "delete-origin"},
	"/Origin/getOriginList":                    {Path: "/TagsOrigin", IDField: "", Resource: "origin", Action: "get-origin-list"},
	"/Origin/saveOrigin":                       {Path: "/TagsOrigin", IDField: "", Resource: "origin", Action: "save-origin"},
	"/Subject/changeStatus":                    {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "change-status"},
	"/Subject/deleteDistributionFile":          {Path: "/SubjectEdit/:id", IDField: "", Resource: "subject", Action: "delete-distribution-file"},
	"/Subject/deleteTestdataFile":              {Path: "/SubjectEdit/:id", IDField: "", Resource: "subject", Action: "delete-testdata-file"},
	"/Subject/getCondition":                    {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-condition"},
	"/Subject/getConfigYaml":                   {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-config-yaml"},
	"/Subject/getDistributionFiles":            {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-distribution-files"},
	"/Subject/getSubject":                      {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-subject"},
	"/Subject/getSubjectForManager":            {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-subject-for-manager"},
	"/Subject/getSubjectList":                  {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-subject-list"},
	"/Subject/getSubjectListNew":               {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "get-subject-list-new"},
	"/Subject/nextPid":                         {Path: "/SubjectManage", IDField: "", Resource: "subject", Action: "next-pid"},
	"/Subject/subjectSave":                     {Path: "/SubjectEdit/:id", IDField: "", Resource: "subject", Action: "subject-save"},
	"/Tag/deleteTag":                           {Path: "/TagsManage", IDField: "", Resource: "tag", Action: "delete-tag"},
	"/Tag/getTagList":                          {Path: "/TagsManage", IDField: "", Resource: "tag", Action: "get-tag-list"},
	"/Tag/saveTag":                             {Path: "/TagsManage", IDField: "", Resource: "tag", Action: "save-tag"},
	"/Tag/showTagList":                         {Path: "/TagsManage", IDField: "", Resource: "tag", Action: "show-tag-list"},
	"/TestLog/TestLogDetail":                   {Path: "/MyTestLogDesc/:id", IDField: "F_ID", Resource: "testlog", Action: "test-log-detail"},
	"/TestLog/TestPoint":                       {Path: "/MyTestLog", IDField: "", Resource: "testlog", Action: "test-point"},
	"/TestLog/doTest":                          {Path: "/MyTestLogAnswer/:id", IDField: "F_SubjectID", Resource: "testlog", Action: "do-test"},
	"/TestLog/getCodeContent":                  {Path: "/MyTestLogAnswer/:id", IDField: "F_ID", Resource: "testlog", Action: "get-code-content"},
	"/TestLog/getSubjectTestLog":               {Path: "/MyTestLog", IDField: "", Resource: "testlog", Action: "get-subject-test-log"},
	"/TestLog/getTestLogList":                  {Path: "/MyTestLog", IDField: "", Resource: "testlog", Action: "get-test-log-list"},
	"/TestLog/submitNoJudge":                   {Path: "/MyTestLogAnswer/:id", IDField: "", Resource: "testlog", Action: "submit-no-judge"},
	"/User/accountPreview":                     {Path: "/UserManage", IDField: "", Resource: "user", Action: "account-preview"},
	"/User/bindPhone":                          {Path: "/UserManage", IDField: "", Resource: "user", Action: "bind-phone"},
	"/User/changePasswordByPhone":              {Path: "/UserManage", IDField: "", Resource: "user", Action: "change-password-by-phone"},
	"/User/checkTeacher":                       {Path: "/UserCheck", IDField: "", Resource: "user", Action: "check-teacher"},
	"/User/deleteUser":                         {Path: "/UserManage", IDField: "", Resource: "user", Action: "delete-user"},
	"/User/editAccount":                        {Path: "/UserManage", IDField: "", Resource: "user", Action: "edit-account"},
	"/User/editPwd":                            {Path: "/UserManage", IDField: "", Resource: "user", Action: "edit-pwd"},
	"/User/getConfig":                          {Path: "/UserManage", IDField: "", Resource: "user", Action: "get-config"},
	"/User/getMasterList":                      {Path: "/UserManage", IDField: "", Resource: "user", Action: "get-master-list"},
	"/User/getNotCheckTeacherList":             {Path: "/UserManage", IDField: "", Resource: "user", Action: "get-not-check-teacher-list"},
	"/User/getStudentListNew":                  {Path: "/UserManage", IDField: "", Resource: "user", Action: "get-student-list-new"},
	"/User/getUserList":                        {Path: "/UserManage", IDField: "", Resource: "user", Action: "get-user-list"},
	"/User/heartbeat":                          {Path: "/UserManage", IDField: "", Resource: "user", Action: "heartbeat"},
	"/User/inviteStudent":                      {Path: "/UserManage", IDField: "", Resource: "user", Action: "invite-student"},
	"/User/login":                              {Path: "/UserManage", IDField: "", Resource: "user", Action: "login"},
	"/User/loginByPhone":                       {Path: "/UserManage", IDField: "", Resource: "user", Action: "login-by-phone"},
	"/User/oneKeyResetPwd":                     {Path: "/UserManage", IDField: "", Resource: "user", Action: "one-key-reset-pwd"},
	"/User/operationInvite":                    {Path: "/UserManage", IDField: "", Resource: "user", Action: "operation-invite"},
	"/User/register":                           {Path: "/UserManage", IDField: "", Resource: "user", Action: "register"},
	"/User/resetPasswordByPhone":               {Path: "/UserManage", IDField: "", Resource: "user", Action: "reset-password-by-phone"},
	"/User/resetPwd":                           {Path: "/UserManage", IDField: "", Resource: "user", Action: "reset-pwd"},
	"/User/searchInvite":                       {Path: "/UserManage", IDField: "", Resource: "user", Action: "search-invite"},
	"/User/sendCodeForBind":                    {Path: "/UserManage", IDField: "", Resource: "user", Action: "send-code-for-bind"},
	"/User/sendCodeForChangePwd":               {Path: "/UserManage", IDField: "", Resource: "user", Action: "send-code-for-change-pwd"},
	"/User/sendCodeForCheck":                   {Path: "/UserCheck", IDField: "", Resource: "user", Action: "send-code-for-check"},
	"/User/sendCodeForLogin":                   {Path: "/UserManage", IDField: "", Resource: "user", Action: "send-code-for-login"},
	"/User/sendCodeForRegister":                {Path: "/UserManage", IDField: "", Resource: "user", Action: "send-code-for-register"},
	"/User/sendCodeForResetPassword":           {Path: "/UserManage", IDField: "", Resource: "user", Action: "send-code-for-reset-password"},
	"/User/studentCreate":                      {Path: "/UserManage", IDField: "", Resource: "user", Action: "student-create"},
	"/User/studentSort":                        {Path: "/StudentRanking", IDField: "", Resource: "user", Action: "student-sort"},
	"/User/studentUpdate":                      {Path: "/UserManage", IDField: "", Resource: "user", Action: "student-update"},
	"/User/userCreate":                         {Path: "/UserManage", IDField: "", Resource: "user", Action: "user-create"},
	"/subject/batchsyncfromhydro":              {Path: "/SubjectManage", IDField: "", Resource: "hydro", Action: "batchsyncfromhydro"},
	"/subject/tongbufromhydrooj":               {Path: "/SubjectManage", IDField: "", Resource: "hydro", Action: "tongbufromhydrooj"},
	"/user/adminIndex":                         {Path: "/Home", IDField: "", Resource: "home", Action: "admin-index"},
	"/user/schoolMasterIndex":                  {Path: "/Home", IDField: "", Resource: "home", Action: "school-master-index"},
	"/user/studentIndex":                       {Path: "/Home", IDField: "", Resource: "home", Action: "student-index"},
	"/user/teacherIndex":                       {Path: "/Home", IDField: "", Resource: "home", Action: "teacher-index"},
}

// WebURL 按接口路径与已提供的参数拼出网页地址。
// 若页面需要 id 而调用方未提供，返回 needID=true，由调用方交互式询问。
func WebURL(epPath string, fields map[string]string) (url string, needID bool) {
	t, ok := WebTargets[epPath]
	if !ok {
		return "", false
	}
	if !containsPlaceholder(t.Path) {
		return t.Path, false
	}
	id := ""
	if t.IDField != "" {
		id = fields[t.IDField]
	}
	if id == "" {
		// 兜底：取第一个看起来像 ID 的字段
		for _, k := range sortedKeys(fields) {
			if len(k) > 2 && (k[len(k)-2:] == "ID") {
				id = fields[k]
				break
			}
		}
	}
	if id == "" {
		return t.Path, true
	}
	return replacePlaceholder(t.Path, id), false
}

func containsPlaceholder(p string) bool {
	return len(p) >= 3 && p[len(p)-3:] == ":id"
}

func replacePlaceholder(p, id string) string {
	if len(p) >= 3 && p[len(p)-3:] == ":id" {
		return p[:len(p)-3] + id
	}
	return p
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
