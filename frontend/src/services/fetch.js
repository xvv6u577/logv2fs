import axios from "axios";

const API_URL = "http://localhost:8079/v1/";

const login = ({email, password}) => {
	return axios
		.post(API_URL + "login", {
			email,
			password,
		})
		.then((response) => {
			return response.data;
		})
		.catch((err) => {
			console.log(err);
		});
};

export { login };
