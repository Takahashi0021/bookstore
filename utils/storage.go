package utils

import (
	"encoding/json"
	"os"
	"sync"

	"bookstore/models"
)

var StorageMu sync.RWMutex

const (
	BooksFile      = "data/books.json"
	AuthorsFile    = "data/authors.json"
	CategoriesFile = "data/categories.json"
	UsersFile      = "data/users.json"
	FavoritesFile  = "data/favorites.json"
)

func SaveBooks(books map[int]models.Book) error {
	StorageMu.Lock()
	defer StorageMu.Unlock()

	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(BooksFile, data, 0644)
}

func LoadBooks() (map[int]models.Book, error) {
	StorageMu.RLock()
	defer StorageMu.RUnlock()

	books := make(map[int]models.Book)

	file, err := os.ReadFile(BooksFile)
	if err != nil {
		if os.IsNotExist(err) {
			return books, nil
		}
		return nil, err
	}

	err = json.Unmarshal(file, &books)
	return books, err
}

func SaveAuthors(authors map[int]models.Author) error {
	StorageMu.Lock()
	defer StorageMu.Unlock()

	data, err := json.MarshalIndent(authors, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(AuthorsFile, data, 0644)
}

func LoadAuthors() (map[int]models.Author, error) {
	StorageMu.RLock()
	defer StorageMu.RUnlock()

	authors := make(map[int]models.Author)

	file, err := os.ReadFile(AuthorsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return authors, nil
		}
		return nil, err
	}

	err = json.Unmarshal(file, &authors)
	return authors, err
}

func SaveCategories(categories map[int]models.Category) error {
	StorageMu.Lock()
	defer StorageMu.Unlock()

	data, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(CategoriesFile, data, 0644)
}

func LoadCategories() (map[int]models.Category, error) {
	StorageMu.RLock()
	defer StorageMu.RUnlock()

	categories := make(map[int]models.Category)

	file, err := os.ReadFile(CategoriesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return categories, nil
		}
		return nil, err
	}

	err = json.Unmarshal(file, &categories)
	return categories, err
}

func SaveUsers(users map[int]models.User) error {
	StorageMu.Lock()
	defer StorageMu.Unlock()

	usersCopy := make(map[int]models.User)
	for id, user := range users {
		userCopy := user
		userCopy.Password = ""
		usersCopy[id] = userCopy
	}

	data, err := json.MarshalIndent(usersCopy, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(UsersFile, data, 0644)
}

func LoadUsers() (map[int]models.User, error) {
	StorageMu.RLock()
	defer StorageMu.RUnlock()

	users := make(map[int]models.User)

	file, err := os.ReadFile(UsersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return users, nil
		}
		return nil, err
	}

	err = json.Unmarshal(file, &users)
	return users, err
}

func SaveFavorites(favorites map[int]map[int]models.Favorite) error {
	StorageMu.Lock()
	defer StorageMu.Unlock()

	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(FavoritesFile, data, 0644)
}

func LoadFavorites() (map[int]map[int]models.Favorite, error) {
	StorageMu.RLock()
	defer StorageMu.RUnlock()

	favorites := make(map[int]map[int]models.Favorite)

	file, err := os.ReadFile(FavoritesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return favorites, nil
		}
		return nil, err
	}

	err = json.Unmarshal(file, &favorites)
	return favorites, err
}
