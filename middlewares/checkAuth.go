package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"whatsapp_file_handling/utils"

	initializers "whatsapp_file_handling/initializers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type User struct {
	ID       uint      `gorm:"primaryKey"`     // Capitalized "ID"
	Cts      time.Time `gorm:"autoCreateTime"` // Created time
	Uts      time.Time `gorm:"autoUpdateTime"` // Updated time
	UserType string    // User type (must be capitalized)
}

func CheckAuth(c *gin.Context) {
	var authHeader = c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": utils.HeaderMissing, "status": http.StatusUnauthorized})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	authToken := strings.Split(authHeader, " ")

	if len(authToken) != 2 || authToken[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": utils.InvalidTokenFormat})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	tokenString := authToken[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_ACCESS_SECRET")), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := claims["id"].(float64)

	var user User
	var result map[string]interface{}

	var test = initializers.DB.Table("users").Where("id = ?", id).Take(&result).Limit(1)
	if test.Error != nil {
		if errors.Is(test.Error, gorm.ErrRecordNotFound) {
			fmt.Println("User not found")
		} else {
			fmt.Println("Database error:", test.Error)
		}
		return
	}
	fmt.Printf("User record: %+v\n", result)

	if err := initializers.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusUnauthorized})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if user.ID == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("user", user)
	c.Next()
}
