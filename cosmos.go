package main

import (
	"context"
	"encoding/json"
	"fmt"



	"github.com/joho/godotenv"


	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

func startCosmos(writeOutput func(msg string)) error {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, proceeding without it")
	}

	const connectionString = "<azure-cosmos-db-nosql-connection-string>"

	clientOptions := azcosmos.ClientOptions{
		EnableContentResponseOnWrite: true,
	}
	
	client, err := azcosmos.NewClientFromConnectionString(connectionString, &clientOptions)
	if err != nil {
		return err
	}

	writeOutput("Current Status:\tStarting...")

	databaseName, found := os.LookupEnv("CONFIGURATION__AZURECOSMOSDB__DATABASENAME")
	if !found {
		databaseName = "cosmicworks"
	}

	database, err := client.NewDatabase(databaseName)
	if err != nil {
		return err
	}

	writeOutput(fmt.Sprintf("Get database:\t%s", database.ID()))

	containerName, found := os.LookupEnv("CONFIGURATION__AZURECOSMOSDB__CONTAINERNAME")
	if !found {
		containerName = "products"
	}

	container, err := database.NewContainer(containerName)
	if err != nil {
		return err
	}

	writeOutput(fmt.Sprintf("Get container:\t%s", container.ID()))

	{
		item := Item{
			Id:        "aaaaaaaa-0000-1111-2222-bbbbbbbbbbbb",
			Category:  "gear-surf-surfboards",
			Name:      "Yamba Surfboard",
			Quantity:  12,
			Price:     850.00,
			Clearance: false,
		}

		partitionKey := azcosmos.NewPartitionKeyString("gear-surf-surfboards")

		context := context.TODO()

		bytes, err := json.Marshal(item)
		if err != nil {
			return err
		}

		response, err := container.UpsertItem(context, partitionKey, bytes, nil)
		if err != nil {
			return err
		}

		if response.RawResponse.StatusCode == 200 || response.RawResponse.StatusCode == 201 {
			created_item := Item{}
			err := json.Unmarshal(response.Value, &created_item)
			if err != nil {
				return err
			}
			writeOutput(fmt.Sprintf("Upserted item:\t%v", created_item))
		}
		writeOutput(fmt.Sprintf("Status code:\t%d", response.RawResponse.StatusCode))
		writeOutput(fmt.Sprintf("Request charge:\t%.2f", response.RequestCharge))
	}

	{
		item := Item{
			Id:        "bbbbbbbb-1111-2222-3333-cccccccccccc",
			Category:  "gear-surf-surfboards",
			Name:      "Kiama Classic Surfboard",
			Quantity:  25,
			Price:     790.00,
			Clearance: true,
		}

		partitionKey := azcosmos.NewPartitionKeyString("gear-surf-surfboards")

		context := context.TODO()

		bytes, err := json.Marshal(item)
		if err != nil {
			return err
		}

		response, err := container.UpsertItem(context, partitionKey, bytes, nil)
		if err != nil {
			return err
		}

		if response.RawResponse.StatusCode == 200 || response.RawResponse.StatusCode == 201 {
			created_item := Item{}
			err := json.Unmarshal(response.Value, &created_item)
			if err != nil {
				return err
			}
			writeOutput(fmt.Sprintf("Upserted item:\t%v", created_item))
		}
		writeOutput(fmt.Sprintf("Status code:\t%d", response.RawResponse.StatusCode))
		writeOutput(fmt.Sprintf("Request charge:\t%.2f", response.RequestCharge))

	}

	{
		partitionKey := azcosmos.NewPartitionKeyString("gear-surf-surfboards")

		context := context.TODO()

		itemId := "aaaaaaaa-0000-1111-2222-bbbbbbbbbbbb"

		response, err := container.ReadItem(context, partitionKey, itemId, nil)
		if err != nil {
			return err
		}

		if response.RawResponse.StatusCode == 200 {
			read_item := Item{}
			err := json.Unmarshal(response.Value, &read_item)
			if err != nil {
				return err
			}

			writeOutput(fmt.Sprintf("Read item id:\t%s", read_item.Id))
			writeOutput(fmt.Sprintf("Read item:\t%v", read_item))
		}

		writeOutput(fmt.Sprintf("Status code:\t%d", response.RawResponse.StatusCode))
		writeOutput(fmt.Sprintf("Request charge:\t%.2f", response.RequestCharge))
	}

	{
		partitionKey := azcosmos.NewPartitionKeyString("gear-surf-surfboards")

		query := "SELECT * FROM products p WHERE p.category = @category"

		queryOptions := azcosmos.QueryOptions{
			QueryParameters: []azcosmos.QueryParameter{
				{Name: "@category", Value: "gear-surf-surfboards"},
			},
		}

		pager := container.NewQueryItemsPager(query, partitionKey, &queryOptions)

		context := context.TODO()

		items := []Item{}

		requestCharge := float32(0)

		for pager.More() {
			response, err := pager.NextPage(context)
			if err != nil {
				return err
			}

			requestCharge += response.RequestCharge

			for _, bytes := range response.Items {
				item := Item{}
				err := json.Unmarshal(bytes, &item)
				if err != nil {
					return err
				}
				items = append(items, item)
			}
		}

		for _, item := range items {
			writeOutput(fmt.Sprintf("Found item:\t%s\t%s", item.Name, item.Id))
		}
		writeOutput(fmt.Sprintf("Request charge:\t%.2f", requestCharge))
	}

	return nil
}
