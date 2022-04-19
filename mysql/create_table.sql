USE unilab;

CREATE TABLE IF NOT EXISTS `oj_user`(
    `user_id` INT(10) UNSIGNED NOT NULL PRIMARY KEY,
    `user_name` VARCHAR(16) NOT NULL DEFAULT 'unknown', 
    `user_real_name` VARCHAR(50) NOT NULL DEFAULT 'unknown',
    `user_email` VARCHAR(255) NOT NULL DEFAULT '',
    `user_git_tsinghua_id` INT UNSIGNED NOT NULL DEFAULT 0,
    `user_last_login_time` DATETIME NOT NULL DEFAULT NOW(),
    `user_type` TINYINT UNSIGNED NOT NULL,
    `user_signup_time` DATETIME NOT NULL DEFAULT NOW(),
    `user_token` VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户信息';

CREATE TABLE IF NOT EXISTS `oj_course`(
    `course_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `course_name` VARCHAR(32) NOT NULL,
    `course_teacher` VARCHAR(32) NOT NULL,
    `course_term` VARCHAR(64) NOT NULL,
    `course_type` TINYINT UNSIGNED NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='课程信息';

CREATE TABLE IF NOT EXISTS `oj_user_course`(
    `course_id` INT UNSIGNED NOT NULL,
    `user_id` INT(10) UNSIGNED NOT NULL,
    `user_type` VARCHAR(255) NOT NULL,
    CONSTRAINT c_oj_user_course_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT c_oj_user_course_2 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户-课程关联表';

CREATE TABLE IF NOT EXISTS `oj_question`(
    `question_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `question_name` VARCHAR(255) NOT NULL,
    `question_tag` VARCHAR(255) NOT NULL,
    `question_creator` INT(10) UNSIGNED NOT NULL,
    `question_score` INT UNSIGNED NOT NULL,
    `question_testcase_num` INT UNSIGNED NOT NULL,
    `question_memory_limit` INT UNSIGNED NOT NULL,
    `question_time_limit` INT UNSIGNED NOT NULL,
    `question_language` VARCHAR(255) NOT NULL,
    `question_compile_options` VARCHAR(255) NOT NULL,
    `question_test_total_num` INT UNSIGNED NOT NULL DEFAULT 0,
    `question_test_ac_num` INT UNSIGNED NOT NULL DEFAULT 0,
    issue_time DATETIME NOT NULL,
    CONSTRAINT c_oj_question FOREIGN KEY (question_creator) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='题目信息';

CREATE TABLE IF NOT EXISTS `oj_homework`(
    `homework_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `homework_name` VARCHAR(255) NOT NULL,
    `homework_begin_time` DATETIME NOT NULL,
    `homework_due_time` DATETIME NOT NULL,
    `homework_description` VARCHAR(255) default '',
    `course_id` INT UNSIGNED NOT NULL,
    CONSTRAINT c_oj_homework_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='作业信息';

CREATE TABLE IF NOT EXISTS `oj_question_homework`(
    `question_id` INT UNSIGNED NOT NULL,
    `homework_id` INT UNSIGNED NOT NULL,
    CONSTRAINT c_oj_question_homework_1 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT c_oj_question_homework_2 FOREIGN KEY (homework_id) REFERENCES oj_homework(homework_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='作业-题目关联表';

CREATE TABLE IF NOT EXISTS `oj_question_course`(
    `question_id` INT UNSIGNED NOT NULL,
    `course_id` INT UNSIGNED NOT NULL,
    CONSTRAINT c_oj_question_course_1 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT c_oj_question_course_2 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE 
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='课程-题目关联表';

CREATE TABLE IF NOT EXISTS `oj_announcement`(
    `announcement_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `announcement_title` VARCHAR(255) NOT NULL,
    `course_id` INT UNSIGNED NOT NULL,
    `issue_time` DATETIME NOT NULL,
    CONSTRAINT c_oj_announcement_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='公告信息';

CREATE TABLE IF NOT EXISTS `oj_test_run`(
    `test_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `test_launch_time` DATETIME NOT NULL,
    
    `course_id` INT UNSIGNED NOT NULL,
    `question_id` INT UNSIGNED NOT NULL,
    `user_id` INT(10) UNSIGNED NOT NULL,

    `language` VARCHAR(255) NOT NULL,
    `save_dir` VARCHAR(255) NOT NULL,
    `compile_result` VARCHAR(1024) NOT NULL DEFAULT '',
    CONSTRAINT c_oj_test_run_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT c_oj_test_run_2 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT c_oj_test_run_3 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='评测信息，每行对应一次提交';

CREATE TABLE IF NOT EXISTS `oj_testcase_run`(
    `testcase_run_id` INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `testcase_rank`   INT UNSIGNED NOT NULL,
    `testcase_run_state` VARCHAR(255) NOT NULL DEFAULT 'Pending',
    `testcase_run_time_elapsed` INT UNSIGNED NOT NULL DEFAULT 0,
    `testcase_run_memory_usage` INT UNSIGNED NOT NULL DEFAULT 0,
    `test_id` INT UNSIGNED NOT NULL,
    CONSTRAINT c_oj_testcase_run_1 FOREIGN KEY (test_id) REFERENCES oj_test_run(test_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='测例信息';

CREATE TABLE IF NOT EXISTS `os_grade`(
    `user_id` INT(10) UNSIGNED NOT NULL,
    `os_grade_id`   INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `branch_name`   VARCHAR(255) NOT NULL,
    -- `grade` INT UNSIGNED NOT NULL,
    -- `total_grade` INT UNSIGNED NOT NULL,
    -- `trace` VARCHAR(1024) NOT NULL,
    CONSTRAINT os_grade_1 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='操作系统章节成绩';

CREATE TABLE IF NOT EXISTS `os_grade_points`(
    `point_id` INT UNSIGNED NOT NULL,
    `grade_id` INT UNSIGNED NOT NULL,
    `point_name` VARCHAR(255) NOT NULL,
    `passed` TINYINT(1) NOT NULL,
    `score` INT UNSIGNED NOT NULL,
    CONSTRAINT os_grade_points_1 FOREIGN KEY (grade_id) REFERENCES os_grade(os_grade_id) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='操作系统章节测试点';

CREATE TABLE IF NOT EXISTS `os_grade_outputs`(
    `grade_id` INT UNSIGNED NOT NULL,
    `type` VARCHAR(255) NOT NULL,
    `alert_class` VARCHAR(255) NOT NULL,
    `message` VARCHAR(255) NOT NULL,
    `content` VARCHAR(255) NOT NULL,
    `expand` TINYINT(1) NOT NULL,
    CONSTRAINT os_grade_outputs_1 FOREIGN KEY (grade_id) REFERENCES os_grade(os_grade_id) ON DELETE CASCADE ON UPDATE CASCADE
)ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8 COMMENT='操作系统章节测试输出';