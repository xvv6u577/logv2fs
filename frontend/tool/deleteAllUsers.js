require("dotenv").config();
const jsonfile = require("jsonfile");
const axios = require("axios").default;

const users = JSON.parse(
	require("fs").readFileSync("./tool/users.json", "utf8")
);

users.forEach((element) => {
	axios
		.get("http://127.0.0.1:8079/v1/" + "deluser/" + element.email)
		.then((res) => {
			console.log(res.data);
		})
		.catch((err) => {
			console.log(err.date);
		});
});
