package database

import (
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


func GetAnnouncementInfo(annoid uint32) (AnnouncementInfo, error) {
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
		logging.Info(err)
		return info, err
	}
	var course_name string
	err = tx.QueryRow("SELECT course_name FROM oj_course WHERE course_id=?;", course_id).Scan(&course_name)
	if err != nil {
		logging.Info(err)
		return info, err
	}
	// read disk
	file_path := setting.CourseRootDir + strconv.FormatUint(uint64(course_id), 10) + "_" + course_name + "/announcements/"+ strconv.FormatUint(uint64(annoid), 10) + "_announcement.md"
	f, err := os.Open(file_path)
	if err != nil {
		logging.Info(err)
		return info, err
	}
	defer f.Close()
 	content, err := ioutil.ReadAll(f)
	if err != nil {
		logging.Info(err)
		return info, err
	}
	info.Content = string(content)
	_ = tx.Commit()
	logging.Info("GetAnnouncementInfo() commit trans action successfully.")
	return info, nil
}


func CheckAnnouncementAccessPermission(annoID, userID uint32) bool {
	var course_id uint32
	err := db.QueryRow("SELECT course_id FROM oj_announcement WHERE announcement_id=?;", annoID).Scan(&course_id)
	if err != nil {
		logging.Info(err)
		return false;
	}
	return CheckCourseAccessPermission(course_id, userID)
}
