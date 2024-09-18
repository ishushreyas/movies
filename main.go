package main

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "encoding/json"
  
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

type JSONResponse struct {
  Message string `json:"message"`
  Response any `json:"response"`
  Code int `json:"code"`
}

type Movie struct {
  Title string `json:"title"`
  Year int `json:"year"`
  Cast []string `json:"cast"`
  Genres []string `json:"genres"`
  Directors []string `json:"directors"`
  Writers []string `json:"writers"`
  Plot string `json:"plot"`
  Runtime string `json:"runtime"`
}

func init() {
  err := connectToMongoDB()
  if err != nil {
    panic("Failed to Connect to database")
  }
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

func connectToMongoDB() error {
    mongoClient, err := mongo.NewClient(options.Client().ApplyURI(loadEnvVar("MONGOURI")))
    if err != nil {
        return err
    }

    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        return err
    }

    //ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        return err
    }
    fmt.Println("Connected to MongoDB")
    return nil
}

func getMovies(w http.ResponseWriter, r *http.Request) {
  var movies []Movie
  
  coll := mongoClient.Database("").Collection("movies")
  
  filter := bson.D{{}}
  cursor, err := coll.Find(context.TODO(), filter)
  if err != nil {
	panic(err)
  }
  if err = cursor.All(context.TODO(), &movies); err != nil {
	panic(err)
  }

  res := JSONResponse{Message: "Message sent successfully", Response: movies, Code: 0}
  sendJSONResponse(w, http.StatusOK, res)
}

func main() {
  mux := http.NewServeMux()
  
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Print("Hello World")
  }
  mux.HandleFunc("/movies", getMovies)
  
  http.ListenAndServe(":8080", mux)
}