package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/mughalaadi/students-api/internal/storage"
	"github.com/mughalaadi/students-api/internal/types"
	"github.com/mughalaadi/students-api/internal/utils/response"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var student types.Student

		slog.Info("creating a student")

		err := json.NewDecoder(r.Body).Decode(&student)

		if errors.Is(err, io.EOF) {
			response.WriteJSON(w, http.StatusBadGateway, response.GenerateErrorResponse(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.GenerateErrorResponse(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		slog.Info("student created", slog.Int64("id", lastId))

		if err != nil {
			response.WriteJSON(w, http.StatusInternalServerError, response.GenerateErrorResponse(err))
			return
		}

		response.WriteJSON(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("getting student by id", slog.String("id", id))
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJSON(w, http.StatusBadRequest, response.GenerateErrorResponse(fmt.Errorf("invalid id: %s", id)))
			return
		}
		student, err := storage.GetStudentById(intId)
		if err != nil {
			response.WriteJSON(w, http.StatusNotFound, response.GenerateErrorResponse(fmt.Errorf("student not found: %s", id)))
			return
		}
		response.WriteJSON(w, http.StatusOK, student)
	}
}
