package database

import(
	// "fmt"
	"unilab-backend/logging"
	"time"
)

type GradeRecord struct{
	Id uint32
	Test_name string
	Tests []Test
	Outputs []Output
}

type Test struct{
	//member definition
	Id int
	Name string
	// Passed bool
	Score int
	Total_score int
}

type Output struct{
	Id int
	Type string
	// Alert_class string
	Message string
	Content string
	// Expand bool
}

func CreateGradeRecord(userid uint32, branch_name string, tests []Test,outputs []Output, test_status string)(error){
	tx,err := db.Begin()
	if err != nil{
		if tx !=nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateOsRecord() begin trans action failed: %v", err)
	}
	result,err := tx.Exec(`INSERT INTO os_grade
		(grade_time)
		VALUES
		(?);
	`,
		// userid,
		time.Now(),
	)
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	gradeID, err := result.LastInsertId()
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	for _,test := range(tests){
		_,err:= tx.Exec(`INSERT INTO os_grade_points
			(point_id,grade_id,point_name,score,total_score)
			VALUES
			(?,?,?,?,?);
		`,
			test.Id,
			gradeID,
			test.Name,
			// test.Passed,
			test.Score,
			test.Total_score,
		)
		if err != nil{
			_ = tx.Rollback()
			logging.Info(err)
			return err
		}
	}
	for _,output := range(outputs){
		_,err:= tx.Exec(`INSERT INTO os_grade_outputs
			(grade_id,output_id,type,message,content)
			VALUES
			(?,?,?,?,?);
		`,
			gradeID,
			output.Id,
			output.Type,
			// output.Alert_class,
			output.Message,
			output.Content,
			// output.Expand,
		)
		if err != nil{
			_ = tx.Rollback()
			logging.Info(err)
			return err
		}
	}
	_,err = tx.Exec(`INSERT IGNORE os_grade_result
		(grade_id,user_id,branch_name,pass_time,total_time)
		VALUES
		(?,?,?,?,?);
	`,
		gradeID,
		userid,
		branch_name,
		0,
		0,
	)
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	if test_status=="passed"{
		_,err = tx.Exec(`UPDATE os_grade_result SET pass_time=pass_time+1,total_time=total_time+1,grade_id=? WHERE user_id=? AND branch_name=?`, gradeID, userid, branch_name)
	}else{
		_,err = tx.Exec(`UPDATE os_grade_result SET total_time=total_time+1 WHERE user_id=? AND branch_name=?`, userid, branch_name)
	}
	if err != nil{
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	_ = tx.Commit()
	logging.Info("CreateSubmitRecord() commit trans action successfully.")
	return nil
}

func GetGradeDetailByBranch(userID uint32,branch_name string) (GradeRecord,error){
	gradeRecord := GradeRecord{}
	tests := []Test{}
	outputs := []Output{}
	err := db.QueryRow("SELECT os_grade_id,branch_name FROM os_grade WHERE user_id=? AND branch_name=?;",userID,branch_name).Scan(
		&gradeRecord.Id,
		&gradeRecord.Test_name,
	)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	point_rows,err := db.Query("SELECT point_id,point_name,score,total_score FROM os_grade_points WHERE grade_id=?;",gradeRecord.Id)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	defer point_rows.Close()
	for point_rows.Next(){
		var test_point Test
		err := point_rows.Scan(&test_point.Id,&test_point.Name,&test_point.Score,&test_point.Total_score)
		if err != nil{
			logging.Info(err)
			continue
		}
		tests = append(tests,test_point)
	}
	output_rows,err := db.Query("SELECT output_id,type,message,content FROM os_grade_outputs WHERE grade_id=?;",gradeRecord.Id)
	if err != nil{
		logging.Info(err)
		return gradeRecord,err
	}
	defer output_rows.Close()
	for output_rows.Next(){
		var output_point Output
		err := output_rows.Scan(&output_point.Id,&output_point.Type,&output_point.Message,&output_point.Content)
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
			logging.Info(err)
		}
		gradeDetails = append(gradeDetails,gradeRecord)
	}
	return gradeDetails,nil
}