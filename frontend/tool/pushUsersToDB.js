require('dotenv').config();
const jsonfile = require("jsonfile");
const axios = require("axios").default;

const users = JSON.parse(require('fs').readFileSync('./tool/users.json', 'utf8'));

users.forEach((element) => {
	console.log(element);
	axios
		.post("http://127.0.0.1:8079/v1/" + "signup", {
			...element,
		})
		.then((res) => {
			console.log(res.data);
		})
		.catch((err) => {
			console.log(err.data);
		});
});

// axios
// 	.post(process.env.REACT_APP_API_HOST +"signup", {
// 		email: "testuser",
// 		path: "ray",
// 		status: "plain",
// 		role: "admin",
// 		password: "testuser",
// 		uuid: "ae4ad192-7776-460b-5b10-646ad2ba3b08",
// 	})
// 	.then((res) => {
// 		console.log(res.data);
// 	})
// 	.catch((err) => {
// 		console.log(err.data);
// 	});
