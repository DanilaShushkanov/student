
?
ss_teacher.protostudent"?
CreateTeacherRequest:
position_type (2.student.PositionTypeRpositionType
	full_name (	RfullName

student_id (R	studentId"?
Teacher
id (Rid:
position_type (2.student.PositionTypeRpositionType
	full_name (	RfullName

student_id (R	studentId"
UpdateTeacherRequest
id (Rid:
position_type (2.student.PositionTypeRpositionType
	full_name (	RfullName"5
ListTeacherRequest
teacher_ids (R
teacherIds"C
ListTeacherResponse,
teachers (2.student.TeacherRteachers*9
PositionType
POSTGRADUATE 
	ASSISTANT
DEAN2?
TeacherServiceB
CreateTeacher.student.CreateTeacherRequest.student.Teacher" A
PatchTeacher.student.UpdateTeacherRequest.student.Teacher" K
ListTeachers.student.ListTeacherRequest.student.ListTeacherResponse" BCZAgithub.com/danilashushkanov/student/pkg/studentServiceApi;studentbproto3
?
ss_student.protostudentss_teacher.proto"?
CreateStudentRequest
	full_name (	RfullName
age (Rage
salary (Rsalary9
teachers (2.student.CreateTeacherRequestRteachers"#
GetStudentRequest
id (Rid"?
Student
id (Rid
fullName (	RfullName
age (Rage
salary (Rsalary,
teachers (2.student.TeacherRteachers"5
ListStudentRequest
student_ids (R
studentIds"C
ListStudentResponse,
students (2.student.StudentRstudents"?
UpdateStudentRequest
id (Rid
	full_name (	RfullName
age (Rage
salary (Rsalary9
teachers (2.student.UpdateTeacherRequestRteachers"
SimpleResponse2?
StudentServiceB
CreateStudent.student.CreateStudentRequest.student.Student" <

GetStudent.student.GetStudentRequest.student.Student" K
ListStudents.student.ListStudentRequest.student.ListStudentResponse" A
PatchStudent.student.UpdateStudentRequest.student.Student" F
DeleteStudent.student.GetStudentRequest.student.SimpleResponse" BCZAgithub.com/danilashushkanov/student/pkg/studentServiceApi;studentbproto3