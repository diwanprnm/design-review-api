package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope adalah SATU-SATUNYA bentuk JSON yang keluar dari API (kecuali 204).
//
// Sukses (201):
//
//	{
//	  "success": true,
//	  "code": "created",
//	  "message": "Designer registered successfully",
//	  "data": { "designer": {...} }
//	}
//
// Gagal (409):
//
//	{
//	  "success": false,
//	  "code": "email_taken",
//	  "message": "email already exist",
//	  "data": null
//	}
//
// Aturan:
//   - success: bool, satu-satunya penanda error/sukses di body.
//     Frontend cukup cek `if (!res.data.success)`.
//   - code: string snake_case, untuk logika mesin (i18n, switch-case di Vue).
//     Jangan taruh angka HTTP di sini (409 sudah ada di HTTP status).
//   - message: kalimat manusia (Inggris), boleh tampil langsung di toast.
//   - data: payload. Bentuk lamanya ({"designer": ...}) dipindah utuh ke
//     dalam data, jadi frontend cuma tambah satu level `.data`.
//     Error validasi field bisa taruh detail di sini, misal
//     {"fields": {"email": "must be a valid email"}}.
//   - 204 No Content tetap tanpa body (aturan HTTP), jangan pakai envelope.
type Envelope struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success kirim response sukses. httpStatus = 200/201, code = "ok"/"created".
func Success(c *gin.Context, httpStatus int, code, message string, data gin.H) {
	c.JSON(httpStatus, Envelope{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Fail kirim response gagal. httpStatus harus cocok dengan jenis error
// (400 validasi, 401 auth, 403 forbidden, 404 not found, 409 konflik, 500 internal).
// data boleh nil; isi gin.H untuk detail tambahan (misal field errors).
func Fail(c *gin.Context, httpStatus int, code, message string, data gin.H) {
	var d any
	if data != nil {
		d = data
	}
	c.JSON(httpStatus, Envelope{
		Success: false,
		Code:    code,
		Message: message,
		Data:    d,
	})
}

// Shortcut yang paling sering dipakai — biar handler tetap 1 baris.

func OK(c *gin.Context, message string, data gin.H) {
	Success(c, http.StatusOK, "ok", message, data)
}

func Created(c *gin.Context, message string, data gin.H) {
	Success(c, http.StatusCreated, "created", message, data)
}

func BadRequest(c *gin.Context, code, message string, data gin.H) {
	Fail(c, http.StatusBadRequest, code, message, data)
}
