package database

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
	"unilab-backend/logging"
	"unilab-backend/setting"
)

type CreateAnnouncementForm struct {
	Title    string `json:"title" form:"title" uri:"title" binding:"required"`
	UserID   uint32 `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseID uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
}

type Announcement struct {
	ID        uint32
	Title     string
	IssueTime time.Time
}

type AnnouncementInfo struct {
	Content   string
	ID        uint32
	Title     string
	IssueTime string
}

func CreateNewAnnouncement(announcementForm CreateAnnouncementForm) (uint32, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateNewAnnouncement() begin trans action failed: ", err)
	}
	// insert new announcement
	result, err := tx.Exec(`INSERT INTO oj_announcement
		(announcement_title, issue_time, course_id)
		VALUES
		(?, ?, ?)
	`,
		announcementForm.Title,
		time.Now(),
		announcementForm.CourseID,
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// get announcement id
	announcementID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// create announcement and user relation
	// get course users
	var userIDs = []uint32{}
	rows, err := tx.Query("SELECT user_id FROM oj_user_course WHERE course_id=?;", announcementForm.CourseID)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	for rows.Next() {
		var userID uint32
		err := rows.Scan(&userID)
		if err != nil {
			_ = tx.Rollback()
			logging.Info(err)
			return 0, err
		}
		userIDs = append(userIDs, userID)
	}
	// create user <-> anno relation
	var insertUserAnno = "INSERT INTO oj_user_announcement(user_id, announcement_id) VALUES "
	for idx, userID := range userIDs {
		if idx < len(userIDs)-1 {
			insertUserAnno += fmt.Sprintf("(%d, %d),", userID, announcementID)
		} else {
			insertUserAnno += fmt.Sprintf("(%d, %d);", userID, announcementID)
		}
	}
	_, err = tx.Exec(insertUserAnno)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateNewAnnouncement() commit trans action successfully. ID: ", announcementID)
	return uint32(announcementID), nil
}

func GetAnnouncementsByCourseID(courseID uint32) ([]Announcement, error) {
	// read oj_announcement
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("GetAnnouncementsByCourseID() begin trans action failed, err: ", err)
		return nil, err
	}
	res, err := tx.Query("SELECT announcement_id, announcement_title, issue_time FROM oj_announcement WHERE course_id=?;", courseID)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	announcements := []Announcement{}
	for res.Next() {
		var announcement Announcement
		err := res.Scan(&announcement.ID, &announcement.Title, &announcement.IssueTime)
		if err != nil {
			_ = tx.Rollback()
			logging.Info(err)
			return nil, err
		}
		announcements = append(announcements, announcement)
	}
	_ = tx.Commit()
	logging.Info("GetAnnouncementsByCourseID() commit trans action successfully.")
	return announcements, nil
}

func GetAnnouncementInfo(annoid, userID uint32) (AnnouncementInfo, error) {
	info := AnnouncementInfo{}
	info.ID = annoid
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("GetAnnouncementInfo() begin trans action failed, err: ", err)
		return info, err
	}
	var courseID uint32
	// read database
	var issueTime time.Time
	err = tx.QueryRow("SELECT announcement_title, course_id, issue_time FROM oj_announcement WHERE announcement_id=?;",
		annoid,
	).Scan(
		&info.Title,
		&courseID,
		&issueTime,
	)
	info.IssueTime = issueTime.Format("2006/01/02 15:04")
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	var courseName string
	err = tx.QueryRow("SELECT course_name FROM oj_course WHERE course_id=?;", courseID).Scan(&courseName)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	// read disk
	filePath := setting.CourseRootDir + strconv.FormatUint(uint64(courseID), 10) + "_" + courseName + "/announcements/" + strconv.FormatUint(uint64(annoid), 10) + "_announcement.md"
	f, err := os.Open(filePath)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	info.Content = string(content)
	_ = tx.Commit()
	logging.Info("GetAnnouncementInfo() commit trans action successfully.")
	// log access
	_, err = db.Exec("UPDATE oj_user_announcement SET access_count=access_count+1 WHERE announcement_id=? AND user_id=?;", annoid, userID)
	if err != nil {
		logging.Info(err)
		return info, err
	}
	return info, nil
}

func CheckAnnouncementAccessPermission(annoID, userID uint32) bool {
	var courseID uint32
	err := db.QueryRow("SELECT course_id FROM oj_announcement WHERE announcement_id=?;", annoID).Scan(&courseID)
	if err != nil {
		logging.Info(err)
		return false
	}
	return CheckCourseAccessPermission(courseID, userID)
}
