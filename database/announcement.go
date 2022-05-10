package database

import (
	"fmt"
	"io/ioutil"
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
	announcement_id, err := result.LastInsertId()
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
	var insertUserAnno string = "INSERT INTO oj_user_announcement(user_id, announcement_id) VALUES "
	for idx, userID := range userIDs {
		if idx < len(userIDs)-1 {
			insertUserAnno += fmt.Sprintf("(%d, %d),", userID, announcement_id)
		} else {
			insertUserAnno += fmt.Sprintf("(%d, %d);", userID, announcement_id)
		}
	}
	_, err = tx.Exec(insertUserAnno)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateNewAnnouncement() commit trans action successfully. ID: ", announcement_id)
	return uint32(announcement_id), nil
}

func GetAnnouncementsByCourseID(course_id uint32) ([]Announcement, error) {
	// read oj_announcement
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("GetAnnouncementsByCourseID() begin trans action failed, err: ", err)
		return nil, err
	}
	res, err := tx.Query("SELECT announcement_id, announcement_title, issue_time FROM oj_announcement WHERE course_id=?;", course_id)
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
	var course_id uint32
	// read database
	var issue_time time.Time
	err = tx.QueryRow("SELECT announcement_title, course_id, issue_time FROM oj_announcement WHERE announcement_id=?;",
		annoid,
	).Scan(
		&info.Title,
		&course_id,
		&issue_time,
	)
	info.IssueTime = issue_time.Format("2006/01/02 15:04")
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	var course_name string
	err = tx.QueryRow("SELECT course_name FROM oj_course WHERE course_id=?;", course_id).Scan(&course_name)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	// read disk
	file_path := setting.CourseRootDir + strconv.FormatUint(uint64(course_id), 10) + "_" + course_name + "/announcements/" + strconv.FormatUint(uint64(annoid), 10) + "_announcement.md"
	f, err := os.Open(file_path)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	info.Content = string(content)
	// log access
	_, err = tx.Exec("UPDATE oj_user_announcement SET access_count=access_count+1 WHERE announcement_id=? AND user_id=?;", annoid, userID)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return info, err
	}
	_ = tx.Commit()
	logging.Info("GetAnnouncementInfo() commit trans action successfully.")
	return info, nil
}

func CheckAnnouncementAccessPermission(annoID, userID uint32) bool {
	var course_id uint32
	err := db.QueryRow("SELECT course_id FROM oj_announcement WHERE announcement_id=?;", annoID).Scan(&course_id)
	if err != nil {
		logging.Info(err)
		return false
	}
	return CheckCourseAccessPermission(course_id, userID)
}
