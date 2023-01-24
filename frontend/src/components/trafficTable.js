import { formatBytes } from "../service/service";

function TrafficTable({ data, limit, by }) {
	return (
		<div className="overflow-x-auto relative shadow-md sm:rounded-lg">
			<table className="table-auto w-full text-sm text-left text-gray-500 dark:text-gray-400">
				<thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
					<tr>
						<th scope="col" className="w-1/5 py-4 px-2">#</th>
						<th scope="col" className="w-1/5 py-4 px-2">{by}</th>
						<th scope="col" className="w-1/5 py-4 px-2">Data Used</th>
						<th scope="col" className="w-2/5 py-4 px-2">By Domain</th>
					</tr>
				</thead>
				<tbody>
					{limit
						? data && data
							.sort((a, b) => b.period - a.period)
							.slice(0, limit)
							.map((item, index) => {
								return (
									<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
										<td className="py-4 px-2">{index + 1}</td>
										<td className="py-4 px-2">{item.period}</td>
										<td className="py-4 px-2">{formatBytes(item.amount)}</td>
										<td className="py-4 px-2">{item.used_by_domain && Object.entries(item.used_by_domain).map(([key, value]) => {
											return (
											<div key={key}>
												<span className="d-block"><span className="inline-block w-44">{key}</span>:{" "}
													<span className="inline-flex justify-center w-24 bg-green-100 text-green-800 text-sm font-medium mr-2 px-0 py-0.5 rounded dark:bg-green-200 dark:text-green-900">
														{formatBytes(value)}
													</span>
												</span>
											</div>)
										})}
										</td>
									</tr>
								);
							})
						: data && data
							.sort((a, b) => b.period - a.period)
							.map((item, index) => {
								return (
									<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
										<td className="py-4 px-2">{index + 1}</td>
										<td className="py-4 px-2">{item.period}</td>
										<td className="py-4 px-2">{formatBytes(item.amount)}</td>
										<td className="py-4 px-2">{item.amount && Object.entries(item.used_by_domain).map(([key, value]) => {
											return (<div key={key}><span className="d-block">{key}:{" "}<span className="bg-green-100 text-green-800 text-sm font-medium mr-2 px-0 py-0.5 rounded dark:bg-green-200 dark:text-green-900">{formatBytes(value)}</span></span><br /></div>)
										})}
										</td>
									</tr>
								);
							})}
				</tbody>
			</table>
		</div>
	);
}

export default TrafficTable;
