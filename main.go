package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Client

type JSONResponse struct {
	Message string `json:"message"`
	Response any `json:"response"`
	Code int `json:"code"`
  }
  
  type Movie struct {
	ID          string    `json:"id" bson:"_id"`
	Plot        string    `json:"plot" bson:"plot"`
	Genres      []string  `json:"genres" bson:"genres"`
	Runtime     int       `json:"runtime" bson:"runtime.$numberInt"`
	Cast        []string  `json:"cast" bson:"cast"`
	Poster      string    `json:"poster" bson:"poster"`
	Title       string    `json:"title" bson:"title"`
	FullPlot    string    `json:"fullplot" bson:"fullplot"`
	Languages   []string  `json:"languages" bson:"languages"`
	Released    int64     `json:"released" bson:"released.$date.$numberLong"`
	Directors   []string  `json:"directors" bson:"directors"`
	Rated       string    `json:"rated" bson:"rated"`
	Awards      Awards    `json:"awards" bson:"awards"`
	LastUpdated string    `json:"lastupdated" bson:"lastupdated"`
	Year        int       `json:"year" bson:"year.$numberInt"`
	IMDB        IMDB      `json:"imdb" bson:"imdb"`
	Countries   []string  `json:"countries" bson:"countries"`
	Type        string    `json:"type" bson:"type"`
	Tomatoes    Tomatoes  `json:"tomatoes" bson:"tomatoes"`
	Comments    int       `json:"num_mflix_comments" bson:"num_mflix_comments.$numberInt"`
}

type Awards struct {
	Wins        int    `json:"wins" bson:"wins.$numberInt"`
	Nominations int    `json:"nominations" bson:"nominations.$numberInt"`
	Text        string `json:"text" bson:"text"`
}

type IMDB struct {
	Rating float64 `json:"rating" bson:"rating.$numberDouble"`
	Votes  int     `json:"votes" bson:"votes.$numberInt"`
	ID     int     `json:"id" bson:"id.$numberInt"`
}

type Tomatoes struct {
	Viewer   Viewer   `json:"viewer" bson:"viewer"`
	Critic   Critic   `json:"critic" bson:"critic"`
	Fresh    int      `json:"fresh" bson:"fresh.$numberInt"`
	Rotten   int      `json:"rotten" bson:"rotten.$numberInt"`
	LastUpdated int64 `json:"lastUpdated" bson:"lastUpdated.$date.$numberLong"`
}

type Viewer struct {
	Rating     float64 `json:"rating" bson:"rating.$numberDouble"`
	NumReviews int     `json:"numReviews" bson:"numReviews.$numberInt"`
	Meter      int     `json:"meter" bson:"meter.$numberInt"`
}

type Critic struct {
	Rating     float64 `json:"rating" bson:"rating.$numberDouble"`
	NumReviews int     `json:"numReviews" bson:"numReviews.$numberInt"`
	Meter      int     `json:"meter" bson:"meter.$numberInt"`
}
  
  func sendJSONResponse(w http.ResponseWriter, s int, r interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	return json.NewEncoder(w).Encode(r)
  }

  func loadEnvVar(v string) string {
	// Load env only in development (optional)
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("env not found in RAILWAY development")
		}
	}

	if os.Getenv("RENDER_SERVICE_ID") == "" { // (Render sets RENDER_SERVICE_ID in production)
        err := godotenv.Load()
        if err != nil {
            log.Println("env not found in RENDER development")
        }
    }

	envVar := os.Getenv(v)
	if envVar == "" {
		log.Fatalf("%s not found in environment variables", v)
	}

	return envVar
}

func connectToMongoDB() *mongo.Client {
    client, err := mongo.NewClient(options.Client().ApplyURI(loadEnvVar("MONGOURI")))
    if err != nil {
        log.Fatal(err)
    }

    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }

    //ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB")
    return client
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	var movies []Movie
	
	// Access the MongoDB collection
	coll := mongoDB.Database("sample_mflix").Collection("movies")
	
	// Get the total count of documents in the collection
	count, err := coll.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		log.Println("Error counting documents:", err)
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}
	
	// Generate a random skip value to fetch random movies
	rand.Seed(time.Now().UnixNano())
	skipValue := rand.Intn(int(count))
	
	// Setup find options to skip and limit the documents
	findOptions := options.Find()
	findOptions.SetSkip(int64(skipValue))
	findOptions.SetLimit(10)
	
	// Find movies with the given options
	cursor, err := coll.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		log.Println("Error fetching movies:", err)
		http.Error(w, "Error fetching movies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())
	
	// Decode the documents into the movies slice
	if err = cursor.All(context.TODO(), &movies); err != nil {
		log.Println("Error decoding movies:", err)
		http.Error(w, "Error decoding movies", http.StatusInternalServerError)
		return
	}
	
	// Check if no movies were found
	if len(movies) == 0 {
		res := JSONResponse{
			Message:  "No movies found",
			Response: nil,
			Code:     1,
		}
		sendJSONResponse(w, http.StatusOK, res)
		return
	}
	
	// Send the successful response with movies
	res := JSONResponse{
		Message:  "Movies fetched successfully",
		Response: movies,
		Code:     0,
	}
	sendJSONResponse(w, http.StatusOK, res)
}

func enableCors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow only requests from frontend 
        var SITE_URL = "http://localhost:5173"
        origin := r.Header.Get("Origin")
        if origin == SITE_URL {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        }

        // Handle preflight requests (OPTIONS)
        if r.Method == "OPTIONS" && origin == SITE_URL {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
	mongoDB = connectToMongoDB()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hey boy")
	})
        mux.HandleFunc("/movies", getMovies)
	http.ListenAndServe(":8080", enableCors(mux))
}
