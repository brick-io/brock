package brock_test

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/gomega"

	"go.onebrick.io/brock"
)

func testSQL(t *testing.T) {
	t.Parallel()

	Expect := NewWithT(t).Expect
	ctx := context.Background()

	db, mock, err := sqlmock.New()
	Expect(err).Should(Succeed())

	mock.MatchExpectationsInOrder(true)

	c := brock.SQL.RoundRobin(db, db, db)

	t.Run("boxexec", func(t *testing.T) {
		query := "INSERT INTO things (name) VALUES ('alpha'),('beta'),('gamma'),('delta')"
		mock.ExpectExec(brock.SQL.Helper.EscapeQuery(query)).WillReturnResult(sqlmock.NewResult(4, 4))

		rowsAffected, lastInsertID := 0, 0
		err = brock.SQL.Box.Exec(c.ExecContext(ctx, query)).Scan(&rowsAffected, &lastInsertID)
		Expect(err).Should(Succeed())
		Expect(rowsAffected).Should(Equal(4))
		Expect(lastInsertID).Should(Equal(4))
	})

	t.Run("boxquery", func(t *testing.T) {
		type resultType1 struct {
			id   []byte
			name string
		}
		var resultList1 []resultType1

		query := "SELECT id, name FROM things"
		mock.ExpectQuery(brock.SQL.Helper.EscapeQuery(query)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow([]byte("id-alpha"), "alpha").
			AddRow([]byte("id-beta"), "beta").
			AddRow([]byte("id-gamma"), "gamma").
			AddRow([]byte("id-gamma"), "delta"),
		)

		err = brock.SQL.Box.Query(c.QueryContext(ctx, query)).Scan(func(i int) []any {
			if i >= 2 {
				return nil
			}
			resultList1 = append(resultList1, resultType1{})

			return []any{
				&resultList1[i].id,
				&resultList1[i].name,
			}
		})
		Expect(err).Should(Succeed())
		Expect(len(resultList1)).Should(Equal(2))
		Expect(resultList1[0]).Should(Equal(resultType1{id: []byte("id-alpha"), name: "alpha"}))
		Expect(resultList1[1]).Should(Equal(resultType1{id: []byte("id-beta"), name: "beta"}))
		Expect(mock.ExpectationsWereMet()).Should(Succeed())
	})
}
