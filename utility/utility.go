package utility

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignedUser struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username,omitempty"`
	Password string             `bson:"password,omitempty"`
}

func Remove(list []string, st string) ([]string, error) {

	index := func(vs []string, t string) int {
		for i, v := range vs {
			if v == t {
				return i
			}
		}
		return -1
	}(list, st)

	if index == -1 {
		return list, errors.New("Cannot find requested object")
	}

	if index == len(list)-1 {
		list = list[0 : len(list)-1]
	} else if index == 0 {
		list = list[1:len(list)]
	} else {
		aux := list[0:index]
		aux2 := list[index+1 : len(list)]
		list = aux
		list = append(list, aux2...)
	}

	return list, nil
}

func SplitDataString(data string) (string, string) {
	userContainer := strings.Split(data, ",")[0]
	userSplitted := strings.Split(userContainer, ":")[1]
	userData := userSplitted[1 : len(userSplitted)-1]

	passContainer := strings.Split(data, ",")[1]
	passSplitted := strings.Split(passContainer, ":")[1]
	passData := passSplitted[1 : len(passSplitted)-2]

	return userData, passData
}

func DataBaseConnection(ctx context.Context, uri string, db string) (*mongo.Client, *mongo.Database) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	database := client.Database(db)

	return client, database
}

func DataBaseCollection(database *mongo.Database, collection string) *mongo.Collection {
	data := database.Collection(collection)
	return data
}

func GetAllCollectionData(ctx context.Context, colct *mongo.Collection, res interface{}) {
	cursor, err := colct.Find(ctx, bson.M{})

	if err != nil {
		panic(err)
	}

	if err = cursor.All(ctx, res); err != nil {
		panic(err)
	}
}

func InsertDataCollection(ctx context.Context, colct *mongo.Collection, data interface{}) error {
	_, err := colct.InsertOne(ctx, data)

	return err
}
