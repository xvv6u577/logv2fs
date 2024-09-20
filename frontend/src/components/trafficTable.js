import { formatBytes } from "../service/service";

function TrafficTable({ data, limit, by }) {

	let sortedData = []

	if (by === "月份") {
		sortedData = data?.sort((a, b) => b.month - a.month)
	} else if (by === "日期") {
		sortedData = data?.sort((a, b) => b.date - a.date).slice(0, 30)
	}

	return (
		<div className="overflow-x-auto relative shadow-md sm:rounded-lg">
			<table className="table-auto w-full text-sm text-left text-gray-500 dark:text-gray-400">
				<thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
					<tr>
						<th scope="col" className="w-1/5 py-4 mx-auto">#</th>
						<th scope="col" className="w-1/5 py-4 px-2">{by}</th>
						<th scope="col" className="w-1/5 py-4 px-2">Data Used</th>
					</tr>
				</thead>
				<tbody>
					{limit
						? sortedData
							?.slice(0, limit)
							.map((item, index) => {
								return (
									<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
										<td className="py-4 px-2">{index + 1}</td>
										<td className="py-4 px-2">{by === "月份" ? item.month : item.date}</td>
										<td className="py-4 px-2">{formatBytes(item.traffic)}</td>
									</tr>
								);
							})
						: sortedData?.map((item, index) => {
								return (
									<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
										<td className="py-4 px-2">{index + 1}</td>
										<td className="py-4 px-2">{by === "月份" ? item.month : item.date}</td>
										<td className="py-4 px-2">{formatBytes(item.traffic)}</td>
									</tr>
								);
							})}
				</tbody>
			</table>
		</div>
	);
}

export default TrafficTable;
