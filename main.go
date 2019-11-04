package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/friendsofgo/graphiql"
	"github.com/graphql-go/graphql"
)

// Job struct
type Job struct {
	ID             int      ` json:"id"`
	Position       string   ` json:position"`
	Company        string   ` json:company"`
	Description    string   ` json:description"`
	SkillsRequired []string ` json:skillsRequired"`
	Location       string   ` json:"location"`
	EmploymentType string   ` json:"employmentType"`
}

type reqBody struct {
	Query string `json:"query"`
}

var jobType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Job",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"position": &graphql.Field{
				Type: graphql.String,
			},
			"company": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"location": &graphql.Field{
				Type: graphql.String,
			},
			"employmentType": &graphql.Field{
				Type: graphql.String,
			},
			"skillsRequired": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
		},
	},
)

func gqlSchema(jobsData []Job) graphql.Schema {
	fields := graphql.Fields{
		"jobs": &graphql.Field{
			Type:        graphql.NewList(jobType),
			Description: "All Jobs",
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return jobsData, nil
			},
		},
		"job": &graphql.Field{
			Type:        jobType,
			Description: "Get Jobs by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, success := params.Args["id"].(int)
				if success {
					for _, job := range jobsData {
						if int(job.ID) == id {
							return job, nil
						}
					}
				}
				return nil, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		fmt.Printf("failed to create new schema, error: %v", err)
	}
	return schema

}

func dataFromJSON() []Job {
	jsonf, err := os.Open("data.json")

	if err != nil {
		fmt.Printf("failed to open json file, error: %v", err)
	}

	jsonDataFromFile, _ := ioutil.ReadAll(jsonf)
	defer jsonf.Close()

	var jobsData []Job

	err = json.Unmarshal(jsonDataFromFile, &jobsData)

	if err != nil {
		fmt.Printf("failed to parse json file, error: %v", err)
	}

	return jobsData

}

func processQuery(query string) (result string) {

	jobsData := dataFromJSON()

	params := graphql.Params{Schema: gqlSchema(jobsData), RequestString: query}

	r := graphql.Do(params)

	if len(r.Errors) > 0 {
		fmt.Printf("Failed to execute query, errors: %+v", r.Errors)
	}

	rJSON, _ := json.Marshal(r)

	return fmt.Sprintf("%s", rJSON)
}

func gqlHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			http.Error(w, "No query data", 400)
			return
		}

		var rBody reqBody
		err := json.NewDecoder(r.Body).Decode(&rBody)
		if err != nil {
			http.Error(w, "Error parsing JSON request body", 400)
		}

		fmt.Fprintf(w, "%s", processQuery(rBody.Query))
	})
}

func main() {

	fmt.Println("Server running in port 3000...")

	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/graphql")
	if err != nil {
		panic(err)
	}

	http.Handle("/graphql", gqlHandler())
	http.Handle("/graphiql", graphiqlHandler)
	http.ListenAndServe(":3000", nil)

}
