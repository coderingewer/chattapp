package models

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Veri tabanı bağlantı linki
const uri = "mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb"

// Değişkenler
var (
	urlDB             *mongo.Database
	userCollection    *mongo.Collection
	messageCollection *mongo.Collection
)

func init() {
	// Veri tabanına bağlanmak için bir istemci oluştur
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	// Birincil geçikme
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	// Veritabanı ve koleksiyonların değişkenlere tanımlanması
	urlDB = client.Database("by37api")
	userCollection = urlDB.Collection("users")
	messageCollection = urlDB.Collection("messages")
	fmt.Println("Successfully connected and pinged.")
}

// Veritabanı ve koleksiyon bağlantısını diğer dosyalarda kullanmak için fonksiyonlar
func getDB() *mongo.Database {
	return urlDB
}

func getUserCollection() *mongo.Collection {
	return userCollection
}

func getMessageCollection() *mongo.Collection {
	return messageCollection
}
