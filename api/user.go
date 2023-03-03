package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

type UserApi struct {
	Database          mongo.Database
	UserCollection    mongo.Collection
	ProfileCollection mongo.Collection
}

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type Profile struct {
	ID                primitive.ObjectID `bson:"_id" json:"_id"`
	Username          string             `bson:"username" json:"username"`
	Email             string             `bson:"email" json:"email"`
	EventParticipated []string           `json:"event_participated" bson:"event_participated"`
	EventHosted       []string           `json:"event_hosted" bson:"event_hosted"`
	EventHistory      []string           `json:"event_history" bson:"event_history"`
	PushToken         string             `json:"push_token" bson:"push_token"`
	HealthStatus      string             `json:"health_status" bson:"health_status"`
	Avatar            string             `json:"avatar" bson:"avatar"`
}

type ProfileUpdateForm struct {
	Email string `json:"email"`
	Query map[string]interface{} `json:"query"`
}

func (u *UserApi) RegisterUser(c *gin.Context) {
	var form User
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	username := form.Username
	password := form.Password
	_, err := u.UserCollection.InsertOne(ctx, bson.M{"email": username, "password": password})
	if err != nil {
		c.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	// generate profile for this user
	var userProfile = Profile{
		Username:          "default",
		Email:             username,
		EventParticipated: []string{},
		EventHosted:       []string{},
		EventHistory:      []string{},
		PushToken:         "",
		HealthStatus:      "negative",
		Avatar:            "",
	}
	r, err := u.ProfileCollection.InsertOne(ctx, userProfile)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err,
		})
		return
	}
	userProfile.ID = r.InsertedID.(primitive.ObjectID)
	c.JSON(200, gin.H{
		"data": userProfile,
	})
}

func (u *UserApi) Login(c *gin.Context) {
	var form User
	var result User
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	username := form.Username
	password := form.Password
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := u.UserCollection.FindOne(ctx, bson.D{{"email",
		username}}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if password != result.Password {
		c.JSON(400, gin.H{
			"error": "incorrect password",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
	})
}

func (u *UserApi) GetProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	email := c.Param("email")
	if email == "" {
		c.JSON(400, gin.H{
			"message": "no email param specified",
		})
		return
	}
	var res Profile
	err := u.ProfileCollection.FindOne(ctx, bson.D{{"email", email}}).Decode(&res)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"data": res,
	})
}

func (u *UserApi) UpdateProfile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var form ProfileUpdateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter := bson.D{{"email", form.Email}}
	var key string
	var value interface{}
	for k, v := range form.Query {
		key = k
		value = v
	}
	update := bson.D{{"$set", bson.D{{key, value}}}}
	result, err := u.ProfileCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(200, gin.H{
			"message": "update fail",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "update success",
	})
}

func (u *UserApi) GetAvatars(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	emailString := c.Query("email")
	var res []string
	emailSegments := strings.Split(emailString, ",")
	for _, email := range emailSegments {
		var profile Profile
		err := u.ProfileCollection.FindOne(ctx, bson.D{{"email", email}}).Decode(&profile)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		res = append(res, profile.Avatar)
	}
	c.JSON(200, gin.H{
		"msg": "success",
		"data": res,
	})
}