package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type EventApi struct {
	Database   mongo.Database
	Collection mongo.Collection
}

type Event struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name         string             `bson:"name" json:"name"`
	Organiser    string             `bson:"organiser" json:"organiser"`
	Preview      string             `bson:"preview" json:"preview"`
	Longitude    float64            `bson:"longitude" json:"longitude"`
	Latitude     float64            `bson:"latitude" json:"latitude"`
	Participants []string           `bson:"participants" json:"participants"`
	Settings     struct {
		Duration       string `bson:"duration" json:"duration"`
		MinParticipant string `bson:"min_participant" json:"min_participant"`
		MaxParticipant string `bson:"max_participant" json:"max_participant"`
		Type           string `bson:"type" json:"type"`
		ThemeColor     string `bson:"theme_color" json:"theme_color"`
		Description    string `bson:"description" json:"description"`
		StartTime      string `bson:"start_time" json:"start_time"`
	} `bson:"settings" json:"settings"`
	Images    []interface{} `bson:"images" json:"images"`
	Active    string        `bson:"active" json:"active"`
	CreatedAt string        `bson:"created_at" json:"created_at"`
}

type AddEventForm struct {
	Name         string   `json:"name"`
	Organiser    string   `json:"organiser"`
	Preview      string   `json:"preview"`
	Longitude    float64  `json:"longitude"`
	Latitude     float64  `json:"latitude"`
	Participants []string `json:"participants"`
	Settings     struct {
		Duration       string `json:"duration"`
		MinParticipant string `json:"min_participant"`
		MaxParticipant string `json:"max_participant"`
		Type           string `json:"type"`
		ThemeColor     string `json:"theme_color"`
		Description    string `json:"description"`
		StartTime      string `json:"start_time"`
	} `json:"settings"`
	Images []interface{} `json:"images"`
}

type EventUpdateForm struct {
	EventId string                 `json:"event_id"`
	Query   map[string]interface{} `json:"query"`
}

type EventDeleteForm struct {
	EventId string `json:"event_id"`
}

func (e *EventApi) GetEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	eventId := c.Query("event_id")
	if eventId == "" {
		events, err := e.GetAllEvents(ctx)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "success",
				"data": events,
			})
		}
	} else {
		event, err := e.GetEventById(ctx, eventId)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"msg":  "success",
				"data": event,
			})
		}
	}
}

func (e *EventApi) GetAllEvents(ctx context.Context) ([]Event, error) {
	collection := e.Collection
	var events []Event
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result Event
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		events = append(events, result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (e *EventApi) GetEventById(ctx context.Context, eventId string) (*Event, error) {
	collection := e.Collection
	var res Event
	Id, err := primitive.ObjectIDFromHex(eventId)
	if err != nil {
		return nil, err
	}
	err = collection.FindOne(ctx, bson.D{{"_id", Id}}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (e *EventApi) AddEvent(c *gin.Context) {
	var form AddEventForm
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	var addedEvent = Event{
		Name:         form.Name,
		Organiser:    form.Organiser,
		Preview:      form.Preview,
		Longitude:    form.Longitude,
		Latitude:     form.Latitude,
		Participants: form.Participants,
		Settings: struct {
			Duration       string `bson:"duration" json:"duration"`
			MinParticipant string `bson:"min_participant" json:"min_participant"`
			MaxParticipant string `bson:"max_participant" json:"max_participant"`
			Type           string `bson:"type" json:"type"`
			ThemeColor     string `bson:"theme_color" json:"theme_color"`
			Description    string `bson:"description" json:"description"`
			StartTime      string `bson:"start_time" json:"start_time"`
		}(form.Settings),
		Images:    form.Images,
		Active:    "false",
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	result, err := e.Collection.InsertOne(ctx, addedEvent)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// convert interface to ObjectID in mongoDB
	Id, _ := primitive.ObjectIDFromHex(result.InsertedID.(primitive.ObjectID).Hex())
	addedEvent.ID = Id
	c.JSON(200, gin.H{
		"msg":  "success",
		"data": addedEvent,
	})
}

func (e *EventApi) UpdateEvent(c *gin.Context) {
	var form EventUpdateForm
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	Id, err := primitive.ObjectIDFromHex(form.EventId)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter := bson.D{{"_id", Id}}
	var key string
	var value interface{}
	for k, v := range form.Query {
		key = k
		value = v
	}
	update := bson.D{{"$set", bson.D{{key, value}}}}
	result, err := e.Collection.UpdateOne(ctx, filter, update)
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

func (e *EventApi) DeleteEvent(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var form EventDeleteForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	event, err := e.GetEventById(ctx, form.EventId)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	organiserEmail := event.Organiser
	// find organiser by email
	var res Profile
	database := e.Database
	profileCollection := database.Collection("Profile")
	err = profileCollection.FindOne(ctx, bson.D{{"email", organiserEmail}}).Decode(&res)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	var newEventHosted, newEventHistory, newEventParticipated []string
	// delete event from organiser's event list
	newEventHosted = deleteFromList(res.EventHosted, form.EventId)
	// delete event from organiser's event history
	newEventHistory = deleteFromList(res.EventHistory, form.EventId)
	// update organiser's event list
	filter := bson.D{{"email", organiserEmail}}
	update := bson.D{{"$set", bson.D{{"event_hosted", newEventHosted}, {"event_history", newEventHistory}}}}
	_, err = profileCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// for each participant, delete event from their event history and event participated
	participants := event.Participants
	for _, participant := range participants {
		var res Profile
		err = profileCollection.FindOne(ctx, bson.D{{"email", participant}}).Decode(&res)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		newEventHistory = deleteFromList(res.EventHistory, form.EventId)
		newEventParticipated = deleteFromList(res.EventParticipated, form.EventId)
		filter := bson.D{{"email", participant}}
		update := bson.D{{"$set", bson.D{{"event_history", newEventHistory}, {"event_participated", newEventParticipated}}}}
		_, err = profileCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
	// delete this event from this collection
	Id, err := primitive.ObjectIDFromHex(form.EventId)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter = bson.D{{"_id", Id}}
	_, err = e.Collection.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "delete success",
		"data": form.EventId,
	})
}

func deleteFromList(list []string, item string) []string {
	for i, v := range list {
		if v == item {
			list = append(list[:i], list[i+1:]...)
			break
		}
	}
	return list
}

func (e *EventApi) GetChats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	chatInfo := c.PostForm("chat_info")
	Id, err := primitive.ObjectIDFromHex(c.PostForm("event_id"))
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	filter := bson.D{{"_id", Id}}
	update := bson.D{{"$set", bson.D{{"chat", chatInfo}}}}
	result, err := e.Collection.UpdateOne(ctx, filter, update)
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