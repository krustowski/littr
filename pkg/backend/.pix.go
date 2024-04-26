package backend

func PixHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}

	// prepare the Logger instance
	l := Logger{
		CallerID:   r.Header.Get("X-API-Caller-ID"),
		IPAddress:  r.Header.Get("X-Real-IP"),
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "pix",
		Version:    r.Header.Get("X-App-Version"),
	}

	request := struct {
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}{}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	err = json.Unmarshal(data, &request)
	if err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	postContent := "/opt/pix/thumb_" + request.Content

	var buffer []byte

	if buffer, err = os.ReadFile(postContent); err != nil {
		resp.Message = err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//compBuff, _ := compressImage(buffer)

	//resp.Data = compBuff
	resp.Data = buffer
	resp.WritePix(w)

	return
}
