package services

import (
	"context"
	"log"
	"time"

	mongo "github.com/erdemkosk/go-config-service/api/db"
	models "github.com/erdemkosk/go-config-service/api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetConfig(types string, key string) (models.Config, error) {
	result := models.Config{}

	filter := bson.M{"type": types, "key": key}

	cachedConfig, cachedError := getCachedConfig(key)

	log.Println(cachedConfig)

	if cachedError == nil && cachedConfig.Key != "" && cachedConfig.Type == types {
		return cachedConfig, cachedError
	}

	client, err := mongo.GetMongoClient()
	if err != nil {
		return result, err
	}

	collection := client.Database(mongo.DB).Collection(mongo.COLLECTION)

	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func CreateConfig(config models.Config) (models.Config, error) {
	client, err := mongo.GetMongoClient()
	if err != nil {
		return models.Config{}, err
	}

	result := models.Config{}
	collection := client.Database(mongo.DB).Collection(mongo.COLLECTION)
	insertResult, err := collection.InsertOne(context.TODO(), config)

	if err != nil {
		return models.Config{}, err
	}

	if err = collection.FindOne(context.TODO(), bson.M{"_id": insertResult.InsertedID}).Decode(&result); err != nil {
		return models.Config{}, err
	}

	setCachedConfig(result.Key, &result)

	return result, err
}

func UpdateConfig(key string, config models.Config) (models.Config, error) {
	client, err := mongo.GetMongoClient()
	if err != nil {
		return models.Config{}, err
	}

	result := models.Config{}

	collection := client.Database(mongo.DB).Collection(mongo.COLLECTION)

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"key": key},
		bson.D{
			{"$set", bson.D{primitive.E{"updated_at", time.Now()}, {"key", config.Key}, {"type", config.Type}, {"value", config.Value}}},
		},
		&opt,
	).Decode(&result)

	deleteCachedConfig(key)
	setCachedConfig(result.Key, &result)

	return result, err
}

func DeleteConfig(key string) (bool, error) {
	client, err := mongo.GetMongoClient()

	var success = false

	if err != nil {
		return success, err
	}

	collection := client.Database(mongo.DB).Collection(mongo.COLLECTION)

	result, err := collection.DeleteOne(
		context.Background(),
		bson.M{"key": key},
	)

	if result.DeletedCount > 0 {
		success = true
	}

	deleteCachedConfig(key)

	return success, err
}
