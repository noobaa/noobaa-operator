package nb

/*func TestCreateSystemAPI(t *testing.T) {
	api := CreateSystemAPI{
		Name:     "name.test",
		Email:    "email@test.io",
		Password: "password!test",
	}
	expectedReq := fmt.Sprintf(`{"api":"system_api","method":"create_system","params":{"name":"%s","email":"%s","password":"%s"}}`, api.Name, api.Email, api.Password)
	expectedRes := fmt.Sprintf(`{"op":"res","reqid":"1","took":11.123,"reply":{"token":"abc"}}`)
	reqBody, res := api.build()
	reqBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)
	assert.Equal(t, expectedReq, string(reqBytes))
	err = json.Unmarshal([]byte(expectedRes), res)
	assert.NoError(t, err)
	assert.Equal(t, api.Response.Op, "res")
	assert.Equal(t, api.Response.RequestID)
	assert.Equal(t, api.Response.Reply.Token, "abc")
}*/
