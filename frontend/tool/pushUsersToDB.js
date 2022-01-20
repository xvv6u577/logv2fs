const jsonfile = require("jsonfile");
const axios = require("axios").default;

const users = jsonfile.readFileSync("tool/users.json");

users.forEach((element) => {
	axios
		.post("http://127.0.0.1:8079/v1/signup", {
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
// 	.post("http://127.0.0.1:8079/v1/signup", {
// 		email: "caster",
// 		path: "ray",
// 		status: "plain",
// 		role: "admin",
// 		password: "caster",
// 		uuid: "ae4ad192-7776-460b-5b10-646ad2ba3b08",
// 	})
// 	.then((res) => {
// 		console.log(res.data);
// 	})
// 	.catch((err) => {
// 		console.log(err.data);
// 	});
