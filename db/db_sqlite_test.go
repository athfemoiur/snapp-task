package db_test

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	. "snapp-task/db"
)

var _ = Describe("SQLiteDB", func() {
	var (
		db             *SQLiteDB
		dataSourceName string
	)

	BeforeEach(func() {
		file, err := os.CreateTemp("", "testdb_*.db")
		Expect(err).NotTo(HaveOccurred())
		dataSourceName = file.Name()

		db, err = NewSQLiteDB(dataSourceName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
		Expect(os.Remove(dataSourceName)).To(Succeed())
	})

	Describe("NewSQLiteDB", func() {
		It("should create a new SQLiteDB instance and create the necessary table", func() {
			var rowCount int
			err := db.Conn.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='matches';").Scan(&rowCount)
			Expect(err).NotTo(HaveOccurred())
			Expect(rowCount).To(Equal(1))
		})
	})

	Describe("SaveData", func() {
		It("should insert data into the matches table", func() {
			err := db.SaveData("http://example.com", "testpattern", "testdata")
			Expect(err).NotTo(HaveOccurred())

			var count int
			query := "SELECT count(*) FROM matches WHERE url = ? AND pattern = ? AND data = ?"
			err = db.Conn.QueryRow(query, "http://example.com", "testpattern", "testdata").Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))
		})
	})

	Describe("Close", func() {
		It("should close the database connection", func() {
			Expect(db.Close()).To(Succeed())
			Expect(db.Conn.Ping()).To(MatchError(fmt.Errorf("sql: database is closed")))
		})
	})
})
