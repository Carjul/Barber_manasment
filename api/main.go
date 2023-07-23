package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BarberShop struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name    string             `json:"barber_shop" bson:"barber_shop"`
	Address string             `json:"address" bson:"address"`
	Phone   string             `json:"phone" bson:"phone"`
}

type Service struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name   string             `json:"name,omitempty" bson:"name,omitempty"`
	Price  int                `json:"price,omitempty" bson:"price,omitempty"`
	DurMin int                `json:"duration_minutes,omitempty" bson:"duration_minutes,omitempty"`
}

type ServiceId struct {
	ServiceID primitive.ObjectID `json:"service_id" bson:"service_id"`
	Date      string             `json:"date" bson:"date"`
}

type Person struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name             string             `json:"name" bson:"name"`
	Phone            string             `json:"phone" bson:"phone"`
	Email            string             `json:"email" bson:"email"`
	LastVisit        string             `json:"last_visit" bson:"last_visit"`
	ServicesReceived []ServiceId        `json:"services_received" bson:"services_received"`
}

var client *mongo.Client
var collection *mongo.Collection
var collection1 *mongo.Collection
var collection2 *mongo.Collection

func main() {

	err2 := godotenv.Load()

	if err2 != nil {
		log.Fatal("Error loading .env file")
	}
	if err := initDB(); err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer client.Disconnect(context.Background())

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "API BARBER!") }).Methods("GET")

	r.HandleFunc("/services", getServicesHandler).Methods("GET")
	r.HandleFunc("/services", createServiceHandler).Methods("POST")
	r.HandleFunc("/services/{id}", GetOneServiceEndpoint).Methods("GET")
	r.HandleFunc("/services/{id}", updateServiceHandler).Methods("PUT")
	r.HandleFunc("/services/{id}", deleteServiceHandler).Methods("DELETE")

	r.HandleFunc("/customers", getCustomersHandler).Methods("GET")
	r.HandleFunc("/customers", createCustomersandler).Methods("POST")
	r.HandleFunc("/customers/{id}", GetOneCustomerEndpoint).Methods("GET")
	r.HandleFunc("/customers/{id}", updateCustomersandler).Methods("PUT")
	r.HandleFunc("/customers/{id}", deleteCustomersandler).Methods("DELETE")

	r.HandleFunc("/datos", getDatosHandler).Methods("GET")
	r.HandleFunc("/datos", createDatosHandler).Methods("POST")
	r.HandleFunc("/datos/{id}", GetOneDatoEndpoint).Methods("GET")
	r.HandleFunc("/datos/{id}", updateDatosHandler).Methods("PUT")
	r.HandleFunc("/datos/{id}", deleteDatosHandler).Methods("DELETE")

	port := os.Getenv("PORT")
	fmt.Println("Server listening on port", port)
	log.Fatal(http.ListenAndServe(port, r))
}

func initDB() error {

	uri := os.Getenv("MONGODB_URI")

	if uri == "" {
		return fmt.Errorf("you must set your 'MONGODB_URI' environmental variable")
	}

	clientOptions := options.Client().ApplyURI(uri)

	var err error

	client, err = mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		return err
	}
	collection = client.Database("barberia").Collection("services")

	collection1 = client.Database("barberia").Collection("customers")

	collection2 = client.Database("barberia").Collection("Datos")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Connected to MongoDB!")

	return client.Ping(ctx, nil)
}

func GetOneServiceEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var service Service
	collection := client.Database("barberia").Collection("services")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, Service{ID: id}).Decode(&service)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(service)
}

func getServicesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var services []Service
	for cursor.Next(ctx) {
		var service Service
		if err := cursor.Decode(&service); err != nil {
			http.Error(w, "Error decoding the data", http.StatusInternalServerError)
			return
		}
		services = append(services, service)
	}

	jsonData, err := json.MarshalIndent(services, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func createServiceHandler(w http.ResponseWriter, r *http.Request) {
	var newService Service
	err := json.NewDecoder(r.Body).Decode(&newService)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, newService)
	if err != nil {
		http.Error(w, "Error creating the service", http.StatusInternalServerError)
		return
	}

	newID := result.InsertedID.(primitive.ObjectID)
	newService.ID = newID

	jsonData, err := json.MarshalIndent(newService, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func updateServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var updatedService Service
	err = json.NewDecoder(r.Body).Decode(&updatedService)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"name":             updatedService.Name,
		"price":            updatedService.Price,
		"duration_minutes": updatedService.DurMin,
	}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error updating the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		http.Error(w, "Error deleting the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getCustomersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection1.Find(ctx, bson.D{})
	if err != nil {
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var customers []Person
	for cursor.Next(ctx) {
		var customer Person
		if err := cursor.Decode(&customer); err != nil {
			log.Println("Error decoding the data:", err)
			http.Error(w, "Error decoding the data", http.StatusInternalServerError)
			return
		}
		customers = append(customers, customer)
	}

	jsonData, err := json.MarshalIndent(customers, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func GetOneCustomerEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "message": "Invalid ID format" }`))
		return
	}
	var customer Person
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = collection1.FindOne(ctx, bson.M{"_id": id}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			response.WriteHeader(http.StatusNotFound)
			response.Write([]byte(`{ "message": "Customer not found" }`))
			return
		}
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "Error querying the database" }`))
		return
	}
	json.NewEncoder(response).Encode(customer)
}

func createCustomersandler(w http.ResponseWriter, r *http.Request) {
	var newCustomer Person
	err := json.NewDecoder(r.Body).Decode(&newCustomer)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection1.InsertOne(ctx, newCustomer)
	if err != nil {
		http.Error(w, "Error creating the service", http.StatusInternalServerError)
		return
	}

	newID := result.InsertedID.(primitive.ObjectID)
	newCustomer.ID = newID

	jsonData, err := json.MarshalIndent(newCustomer, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func updateCustomersandler(w http.ResponseWriter, r *http.Request) {
	customerID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}
	var updatedCustomer Person
	err = json.NewDecoder(r.Body).Decode(&updatedCustomer)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"name":              updatedCustomer.Name,
		"phone":             updatedCustomer.Phone,
		"email":             updatedCustomer.Email,
		"last_visit":        updatedCustomer.LastVisit,
		"services_received": updatedCustomer.ServicesReceived,
	}}

	_, err = collection1.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error updating the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "message": "Customer updated successfully" }`))

}

func deleteCustomersandler(w http.ResponseWriter, r *http.Request) {
	customerID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	_, err = collection1.DeleteOne(ctx, filter)
	if err != nil {
		http.Error(w, "Error deleting the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "message": "Customer deleted successfully" }`))

}
func getDatosHandler(w http.ResponseWriter, r *http.Request) {
	var datos []BarberShop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection2.Find(ctx, bson.D{})
	if err != nil {
		http.Error(w, "Error querying the database", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var dato BarberShop
		if err := cursor.Decode(&dato); err != nil {
			http.Error(w, "Error decoding the data", http.StatusInternalServerError)
			return
		}
		datos = append(datos, dato)
	}
	jsonData, err := json.MarshalIndent(datos, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}

func createDatosHandler(w http.ResponseWriter, r *http.Request) {
	var newDato BarberShop
	err := json.NewDecoder(r.Body).Decode(&newDato)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection2.InsertOne(ctx, newDato)
	if err != nil {
		http.Error(w, "Error creating the service", http.StatusInternalServerError)
		return
	}

	newID := result.InsertedID.(primitive.ObjectID)
	newDato.ID = newID

	jsonData, err := json.MarshalIndent(newDato, "", "    ")
	if err != nil {
		http.Error(w, "Error encoding the data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}

func GetOneDatoEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var dato BarberShop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := collection2.FindOne(ctx, bson.M{"_id": id}).Decode(&dato)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(dato)
}

func updateDatosHandler(w http.ResponseWriter, r *http.Request) {
	datoID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(datoID)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	var updatedDato BarberShop
	err = json.NewDecoder(r.Body).Decode(&updatedDato)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"barber_shop": updatedDato.Name,
		"address":     updatedDato.Address,
		"phone":       updatedDato.Phone,
	}}

	_, err = collection2.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error updating the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "message": "Customer updated successfully" }`))

}

func deleteDatosHandler(w http.ResponseWriter, r *http.Request) {

	datoID := mux.Vars(r)["id"]

	objID, err := primitive.ObjectIDFromHex(datoID)

	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}

	_, err = collection2.DeleteOne(ctx, filter)

	if err != nil {
		http.Error(w, "Error deleting the service", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{ "message": "Customer deleted successfully" }`))

}
