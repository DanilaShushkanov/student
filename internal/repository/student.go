package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/danilashushkanov/student/internal/model"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

const (
	studentTName = "student"
)

type StudentRepositoryImpl struct {
	db                *sqlx.DB
	teacherRepository TeacherRepository
}

func NewStudentRepository(db *sqlx.DB, teacherRepository TeacherRepository) *StudentRepositoryImpl {
	return &StudentRepositoryImpl{
		db:                db,
		teacherRepository: teacherRepository,
	}
}

func (s *StudentRepositoryImpl) Create(ctx context.Context, student *model.Student) (*model.Student, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.WithContext(ctx).WithField("err", rollbackErr).Error("не удалось выполнить Rollback Student")
			}
			return
		}
	}()

	query, _, err := goqu.Insert(studentTName).Rows(student).Returning("id").ToSQL()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Insert Student: %w", err)
	}

	var id int64
	if err = tx.QueryRowxContext(ctx, query).Scan(&id); err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос на создание student: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось выполнить commit: %w", err)
	}

	student.ID = id

	if err = s.createNestedObjects(ctx, student); err != nil {
		return nil, fmt.Errorf("не удалось добавить вложенные объекты: %w", err)
	}

	student, err = s.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить student после доабвления: %w", err)
	}

	return student, nil
}

func (s *StudentRepositoryImpl) createNestedObjects(ctx context.Context, student *model.Student) error {
	for _, teacher := range student.Teachers {
		teacher.StudentID = student.ID
	}

	if _, err := s.teacherRepository.Create(ctx, student.Teachers); err != nil {
		return fmt.Errorf("не удалось добавить вложенные объекты :%w", err)
	}

	return nil
}

type StudentListFilter struct {
	IDList []int64
}

func (f *StudentListFilter) toDataSet() *goqu.SelectDataset {
	selectDataset := goqu.From(studentTName)
	if f.IDList == nil {
		return selectDataset
	}

	selectDataset = selectDataset.Where(
		goqu.I("id").Is(f.IDList),
	)

	return selectDataset
}

func (s *StudentRepositoryImpl) List(ctx context.Context, filter *StudentListFilter) ([]*model.Student, error) {
	studentList := make([]*model.Student, 0)

	query, _, err := filter.toDataSet().ToSQL()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос list student: %w", err)
	}

	if err = s.db.SelectContext(ctx, &studentList, query); err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос list student %w: ", err)
	}

	if err = s.loadNestedObjects(ctx, studentList); err != nil {
		return nil, fmt.Errorf("ошибкак при загрузке вложенных объектов: %w", err)
	}

	return studentList, nil
}

func (s *StudentRepositoryImpl) Get(ctx context.Context, studentID int64) (*model.Student, error) {
	query, _, err := goqu.From(studentTName).Where(
		goqu.I("id").Eq(studentID),
	).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать запрос get Student: %w", err)
	}

	student := &model.Student{}
	if err = s.db.GetContext(ctx, student, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEntityNotFound
		}
		return nil, fmt.Errorf("не удалось выполнить запрос get Student : %w", err)
	}
	err = s.loadNestedObjects(ctx, []*model.Student{student})
	if err != nil {
		return nil, fmt.Errorf("ошибка при загрузке вложенных объектов: %w", err)
	}

	return student, nil
}

func (s *StudentRepositoryImpl) loadNestedObjects(ctx context.Context, students []*model.Student) error {
	if len(students) == 0 {
		return nil
	}

	if err := s.fillTeachers(ctx, students); err != nil {
		return fmt.Errorf("ошибка при загрузке teachers: %w", err)
	}
	return nil
}

func (s *StudentRepositoryImpl) fillTeachers(ctx context.Context, students []*model.Student) error {
	studentIds := make([]int64, 0, len(students))
	studentMap := make(map[int64]*model.Student)
	for _, student := range students {
		studentIds = append(studentIds, student.ID)
		studentMap[student.ID] = student
	}

	teacherListFilter := &TeacherListFilter{IDList: studentIds, FieldName: "student_id"}
	teacherList, err := s.teacherRepository.List(ctx, teacherListFilter)
	if err != nil {
		return fmt.Errorf("ошибка при получении teacher для student: %w", err)
	}

	for _, teacher := range teacherList {
		student, ok := studentMap[teacher.StudentID]
		if ok {
			student.Teachers = append(student.Teachers, teacher)
		}
	}
	return nil
}

func (s *StudentRepositoryImpl) Update(ctx context.Context, student *model.Student) (*model.Student, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.WithContext(ctx).WithField("err", rollbackErr).Error("не удалось выполнить Rollback Update Student")
			}
			return
		}
	}()

	query, _, err := goqu.Update(studentTName).Set(
		student,
	).Where(
		goqu.I("id").Eq(student.ID),
	).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать Update Student: %w", err)
	}

	if _, err = tx.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос по обновлению student: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("не удалось выполнить commit: %w", err)
	}

	if err = s.updateNestedObjects(ctx, student); err != nil {
		return nil, fmt.Errorf("не удалось обновить вложенные объекты: %w", err)
	}

	student, err = s.Get(ctx, student.ID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить student после обновления: %w", err)
	}
	return student, nil
}

func (s *StudentRepositoryImpl) updateNestedObjects(ctx context.Context, student *model.Student) error {
	for _, teacher := range student.Teachers {
		teacher.StudentID = student.ID
	}

	_, err := s.teacherRepository.UpdateTeachers(ctx, student.ID, student.Teachers)
	if err != nil {
		return fmt.Errorf("не удалось обновить вложенные объекты: %w", err)
	}

	return nil
}

func (s *StudentRepositoryImpl) Delete(ctx context.Context, studentID int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.WithContext(ctx).WithField("err", rollbackErr).Error("не удалось выполнить Rollback Student")
			}
			return
		}
	}()

	query, _, err := goqu.Delete(studentTName).Where(
		goqu.I("id").Eq(studentID)).ToSQL()
	if err != nil {
		return fmt.Errorf("не удалось создать Delete Student: %w", err)
	}

	if _, err = tx.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("не удалось выполнить запрос по удалению student: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("не удалось выполнить commit: %w", err)
	}

	return nil
}
