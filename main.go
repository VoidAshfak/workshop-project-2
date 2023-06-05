package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	echo "github.com/labstack/echo/v4"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// DB connection
	var err error
	dsn := "root:ashfak@tcp(localhost:3306)/todo-point?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect to the database!")
	}

	// Create a new instance of the Echo application
	e := echo.New()
	fmt.Println(db)

	// Define routes
	e.GET("/hello", hello)
	e.GET("/user", getUser)
	e.GET("/activity", getActivity)

	// Start the server
	err = e.Start(":8080")
	if err != nil {
		panic(err)
	}
}

// Handler for the "/hello" route
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

type User struct {
	ID             uint   `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
}

// Handler for the "/user" route
func getUser(c echo.Context) error {
	idStr := c.QueryParam("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	var user User
	result := db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

type UserActivity struct {
	ID             uint   `json:"id"`
	FirstName      string `json:"first_name"`
	Country        string `json:"country"`
	ProfilePicture string `json:"profile_picture"`
	Points         uint   `json:"points"`
	Rank           int    `json:"rank"`
}

// Handler for the "/activity" route
func getActivity(c echo.Context) error {
	var userActivities []UserActivity

	// Retrieve users and their activities from the database
	result := db.Table("users").
		Select("users.id, users.first_name, users.country, users.profile_picture, activities.points").
		Joins("JOIN activity_logs ON users.id = activity_logs.user_id").
		Joins("JOIN activities ON activity_logs.activity_id = activities.id").
		Scan(&userActivities)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
	}

	// Calculate points for each user
	pointsMap := calculatePoints(userActivities)

	// Assign points to the user activities
	for i := range userActivities {
		userActivities[i].Points = pointsMap[userActivities[i].ID]
	}

	// Remove duplicate user activities
	uniqueUsers := removeDuplicates(userActivities)

	// Sort user activities based on points in descending order
	sort.SliceStable(uniqueUsers, func(i, j int) bool {
		return uniqueUsers[i].Points > uniqueUsers[j].Points
	})

	// Rank the sorted user activities
	rank := 1
	uniqueUsers[0].Rank = rank
	for i := 1; i < len(uniqueUsers); i++ {
		if uniqueUsers[i].Points < uniqueUsers[i-1].Points {
			rank++
		}
		uniqueUsers[i].Rank = rank
	}

	return c.JSON(http.StatusOK, uniqueUsers)
}

// CalculatePoints calculates the total points for each user
func calculatePoints(users []UserActivity) map[uint]uint {
	pointsMap := make(map[uint]uint)
	for _, user := range users {
		pointsMap[user.ID] += user.Points
	}
	return pointsMap
}

// RemoveDuplicates removes duplicate user activities and returns unique users
func removeDuplicates(users []UserActivity) []UserActivity {
	uniqueUsers := make(map[uint]UserActivity)
	for _, user := range users {
		uniqueUsers[user.ID] = user
	}

	result := make([]UserActivity, 0, len(uniqueUsers))
	for _, user := range uniqueUsers {
		result = append(result, user)
	}
	return result
}
