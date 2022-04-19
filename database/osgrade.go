package database

import(
	"fmt"
	"unilab-backend/logging"
	"unilab-backend/os"
)

type GradeInfo struct{
	Branch_name string
	Grade uint32
	Total_grade uint32
	Trace string
}

func CreateGradeRecord(userid uint32, branch_name string, tests []os.Test,outputs []os.output)(uint32,error){
	tx,err := db.Begin()
	if err != nil{
		if tx !=nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateOsRecord() begin trans action failed: %v", err)
	}
	result,err := tx.Exec(`INSERT INTO os_grade
		(user_id,branch_name,grade,total_grade,trace)
		VALUES
		(?,?,?,?,?,?);
	`,
		userid,
		branch_name,
		grade,
		total_grade,
		trace,
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
	for _,output := range(ouputs){
		_,err:= tx.Exec(`INSERT INTO os_grade_outputs
			(grade_id,type,alert_class,message,content,expand)
			VALUES
			(?,?,?,?,?,?);
		`,
			gradeID,
			output.Id,
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

func GetGradeDetailByBranch(branch_name string, userID uint32) (GradeInfo,error){
	var gradeInfo = GradeInfo{}
	err := db.QueryRow("SELECT branch_name,grade,total_grade,trace FROM os_grade WHERE user_id=? AND branch_name=?;",userID,branch_name).Scan(
		&gradeInfo.Branch_name,
		&gradeInfo.Grade,
		&gradeInfo.Total_grade,
		&gradeInfo.Trace,
	)
	if err != nil{
		logging.Info(err)
		return gradeInfo,err
	}
	return gradeInfo,nil
}

func GetGradeDetailsById(userID uint32) ([]GradeInfo,error){
	var gradeDetails = []GradeInfo{}
	rows,err := db.Query("SELECT branch_name,grade,total_grade,trace FROM os_grade WHERE user_id=?;",userID)
	if err != nil{
		logging.Info(err)
		return gradeDetails,err
	}
	defer rows.Close()
	for rows.Next(){
		var gradeInfo GradeInfo
		err := rows.Scan(&gradeInfo.Branch_name,&gradeInfo.Grade,&gradeInfo.Total_grade,&gradeInfo.Trace)
		if err != nil{
			logging.Info(err)
			continue
		}
		gradeDetails=append(gradeDetails,gradeInfo)
	}
	return gradeDetails,nil
}