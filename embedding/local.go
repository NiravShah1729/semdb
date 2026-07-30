package embedding

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"fmt"
)

type LocalEmbedder struct {
	url    string
	client *http.Client
}

func NewLocalEmbedder(url string) *LocalEmbedder{
	return &LocalEmbedder{
		url: url,	
		client: &http.Client{},
	}
}

type localEmbedRequest struct {
    Text string `json:"text"`
}

type localEmbedResponse struct {
    Embedding []float32 `json:"embedding"`
}

func (l *LocalEmbedder) Embed(text string) ([]float32,error){
	reqBody, err := json.Marshal(localEmbedRequest{Text: text})
	if err != nil {
		return nil,err
	}

	req,err := http.NewRequest("POST",l.url + "/embed",bytes.NewBuffer(reqBody))

	if err != nil {
		return nil,err
	}

	req.Header.Set("Content-Type", "application/json")

	resp,err := l.client.Do(req)

	if err != nil {
		return nil,err
	}

	body,err := io.ReadAll(resp.Body)

	if err != nil {
		return nil,err
	}

	if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("embedding service error: %s", string(body))
    }

	var parsed localEmbedResponse
	if err := json.Unmarshal(body,&parsed); err != nil {
		return nil,err
	}

	return parsed.Embedding,nil
}