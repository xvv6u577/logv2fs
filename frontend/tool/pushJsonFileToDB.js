"use strict";

require('dotenv').config();
const { MongoClient } = require("mongodb");
const jsonData = JSON.parse(require('fs').readFileSync('./tool/db.json', 'utf8'));

const client = new MongoClient(process.env.URI);

let users = [];
Object.keys(jsonData.users).forEach(function (key) {
	let val = jsonData.users[key];
	users.push(val);
});

const getYearStr = (date) => {
	return date.getFullYear()+"";
};

const getMonthStr = (date) => {
	return (
		date.getFullYear() +
		"" +
		(date.getMonth() < 9 ? "0" + (date.getMonth() + 1) : date.getMonth() + 1)
	);
};

const getDayStr = (date) => {
	return (
		date.getFullYear() +
		"" +
		(date.getMonth() < 9 ? "0" + (date.getMonth() + 1) : date.getMonth() + 1) +
		"" +
		(date.getDate() < 10 ? "0" + date.getDate() : date.getDate())
	);
};

async function run() {
	try {
		await client.connect();

		const database = client.db("logV2rayTrafficDB");
		const dbUsers = database.collection("USERS");

		await Promise.all(
			users.map(async (element) => {
				const currentUserTrafficTable = database.collection(element.email);

				let trafficByHour = [],
					byYear = [],
					byMonth = [],
					byDay = [];
				let currentDayTrffic = 0,
					currentMonthTrffic = 0,
					currentYearTrffic = 0,
					used = 0;
				let currentDay, currentMonth, currentYear;
				Object.keys(jsonData[element.email]).forEach(function (key) {
					let val = jsonData[element.email][key];

					let date = new Date(val.timestamp * 1000);
					let current = new Date();
					// { period: "20210223", amount: 12345 }
					let day = getDayStr(date);
					let month = getMonthStr(date);
					let year = getYearStr(date);
					currentDay = getDayStr(current);
					currentMonth = getMonthStr(current);
					currentYear = getYearStr(current);

					let indayForDay = byDay.findIndex((ele) => ele.period === day);
					if (indayForDay !== -1) {
						byDay[indayForDay].amount =
							byDay[indayForDay].amount + val.uplink + val.downlink;
					} else {
						byDay.push({
							period: day,
							amount: val.uplink + val.downlink,
						});
					}
					if (currentDay === day) {
						currentDayTrffic = currentDayTrffic + val.uplink + val.downlink;
					}

					let indayForMonth = byMonth.findIndex((ele) => ele.period === month);
					if (indayForMonth !== -1) {
						byMonth[indayForMonth].amount =
							byMonth[indayForMonth].amount + val.uplink + val.downlink;
					} else {
						byMonth.push({
							period: month,
							amount: val.uplink + val.downlink,
						});
					}
					if (currentMonth === month) {
						currentMonthTrffic = currentMonthTrffic + val.uplink + val.downlink;
					}

					let indayForYear = byYear.findIndex((ele) => ele.period === year);
					if (indayForYear !== -1) {
						byYear[indayForYear].amount =
							byYear[indayForYear].amount + val.uplink + val.downlink;
					} else {
						byYear.push({
							period: year,
							amount: val.uplink + val.downlink,
						});
					}
					if (currentYear === year) {
						currentYearTrffic = currentYearTrffic + val.uplink + val.downlink;
					}

					trafficByHour.push({
						created_at: date,
						total: val.uplink + val.downlink,
					});
					used = used + val.uplink + val.downlink;
				});

				// console.log(byDay);
				// console.log(byMonth);
				// console.log(byYear);
				// console.log(trafficByHour);

				const updateDoc = {
					$set: {
						used: used,
						used_by_current_year: {
							period: currentYear,
							amount: currentYearTrffic,
						},
						used_by_current_month: {
							period: currentMonth,
							amount: currentMonthTrffic,
						},
						used_by_current_day: {
							period: currentDay,
							amount: currentDayTrffic,
						},
						traffic_by_year: byYear,
						traffic_by_month: byMonth,
						traffic_by_day: byDay,
					},
				};
				const query = { email: element.email };
				const options = { upsert: true };
				const updateOneResult = await dbUsers.updateOne(
					query,
					updateDoc,
					options
				);
				console.log("updateOne: ", updateOneResult);

				if (trafficByHour.length > 0) {
					const insertManyOptions = { ordered: true };
					const insertManyResult = await currentUserTrafficTable.insertMany(
						trafficByHour,
						insertManyOptions
					);
					console.log("insertMany: ", insertManyResult);
				}
			})
		);
	} finally {
		await client.close();
	}
}
run().catch(console.dir);
