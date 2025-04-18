package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	parcel.Number = id // <--- важное исправление

	// get
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, retrievedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if err != nil {
		fmt.Println(err)
		return
	}
	// set address

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	// check
	parcel, err = store.Get(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	require.Equal(t, newAddress, parcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	if err != nil {
		fmt.Println(err)
		return
	}
	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	if err != nil {
		fmt.Println(err)
		return
	}

	// check
	parcel, err = store.Get(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	require.Equal(t, newStatus, parcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		if err != nil {
			fmt.Println(err)
			return
		}
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	if err != nil {
		fmt.Println(err)
		return
	}
	require.Equal(t, len(parcels), len(storedParcels))

	// check
	for _, parcel := range storedParcels {
		storedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, storedParcel.Client, parcel.Client)
		require.Equal(t, storedParcel.Status, parcel.Status)
		require.Equal(t, storedParcel.Address, parcel.Address)
	}
	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка

}
