package database

import(
	"fmt"
	"unilab-backend/logging"
)

type GradeRecord struct{
	Id uint32
	Branch_name string
	Tests []Test
	Outputs []Output
}

type Test struct{
	//member definition
	Id int
	Name string
	Passed bool
	Score int
}

type Output struct{
	Id int
	Type string
	Alert_class string
	Message string
	Content string
	Expand bool
}

func CreateGradeRecord(userid uint32, branch_name string, tests []Test,outputs []Output)(uint32,error){
	tx,err := db.Begin()
	if err != nil{
		if tx !=nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateOsRecord() begin trans action failed: %v", err)
	}
	result,err := tx.Exec(`INSERT INTO os_grade
		(user_id,branch_name)
		VALUES
		(?,?);
	`,
		userid,
		branch_name,
	)
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return 0,err
	}
	gradeID, err := result.LastInsertId()
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	for _,test := range(tests){
		_,err:= tx.Exec(`INSERT INTO os_grade_points
			(point_id,grade_id,point_name,passed,score)
			VALUES
			(?,?,?,?,?);
		`,
			test.Id,
			gradeID,
			test.Name,
			test.Passed,
			test.Score,
		)
		if err != nil{
			_ = tx.Rollback()
			logging.Info(err)
			return 0,err
		}
	}
	for _,output := range(outputs){
		_,err:= tx.Exec(`INSERT INTO os_grade_outputs
			(grade_id,output_id,type,alert_class,message,content,expand)
			VALUES
			(?,?,?,?,?,?,?);
		`,
			gradeID,
			output.Id,
			output.Type,
			output.Alert_class,
			output.Message,
			output.Content,
			output.Expand,
		)
		if err != nil{
			_ = tx.Rollback()
			logging.Info(err)
			return 0,err
		}
	}
	_ = tx.Commit()
	logging.Info("CreateSubmitRecord() commit trans action successfully.")
	return uint32(gradeID),nil
}

func GetGradeDetailByBranch(userID uint32,branch_name string) (GradeRecord,error){
	gradeRecord := GradeRecord{}
	tests := []Test{}
	outputs := []Output{}
	err := db.QueryRow("SELECT os_grade_id,branch_name FROM os_grade WHERE user_id=? AND branch_name=?;",userID,branch_name).Scan(
		&gradeRecord.Id,
		&gradeRecord.Branch_name,
	)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	point_rows,err := db.Query("SELECT point_id,point_name,passed,score FROM os_grade_points WHERE grade_id=?;",gradeRecord.Id)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	defer point_rows.Close()
	for point_rows.Next(){
		var test_point Test
		err := point_rows.Scan(&test_point.Id,&test_point.Name,&test_point.Passed,&test_point.Score)
		if err != nil{
			logging.Info(err)
			continue
		}
		tests = append(tests,test_point)
	}
	output_rows,err := db.Query("SELECT output_id,type,alert_class,message,content,expand FROM os_grade_outputs WHERE grade_id=?;",gradeRecord.Id)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	defer output_rows.Close()
	for output_rows.Next(){
		var output_point Output
		err := output_rows.Scan(&output_point.Id,&output_point.Type,&output_point.Alert_class,&output_point.Message,&output_point.Content,&output_point.Expand)
		if err != nil{
			logging.Info(err)
			continue
		}
		outputs = append(outputs,output_point)
	}
	gradeRecord.Tests=tests
	gradeRecord.Outputs=outputs
	return gradeRecord,nil
}

func GetGradeDetailsById(userID uint32) ([]GradeRecord,error){
	var gradeDetails = []GradeRecord{}
	// chs := [...]string{"ch7"}
	rows,err := db.Query("SELECT branch_name FROM os_grade WHERE user_id=?;",userID)
	if err != nil{
		logging.Info(err)
		return gradeDetails,err
	}
	defer rows.Close()
	for rows.Next(){
		var ch string
		err := rows.Scan(&ch)
		if err != nil{
			logging.Info(err)
			return gradeDetails,err
		}
		// fmt.Println(ch)
		gradeRecord,err:= GetGradeDetailByBranch(userID,ch)
		if err != nil{
			gradeDetails = append(gradeDetails,GradeRecord{0,ch,[]Test{},[]Output{}})
		}
		gradeDetails = append(gradeDetails,gradeRecord)
	}
	fmt.Println(gradeDetails)
	// for _,ch := range(chs){
	// 	gradeRecord,err:= GetGradeDetailByBranch(userID,ch)
	// 	if err != nil{
	// 		gradeDetails = append(gradeDetails,GradeRecord{0,ch,[]Test{},[]Output{}})
	// 	}
	// 	gradeDetails = append(gradeDetails,gradeRecord)
	// }
	return gradeDetails,nil
}