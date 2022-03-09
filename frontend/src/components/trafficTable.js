import { Table } from "react-bootstrap";
import { formatBytes } from "../service/service";

function TrafficTable({ data,limit,by }) {
	return (
		<Table striped bordered hover size="sm" variant="dark" className="mx-auto">
			<thead>
				<tr>
					<th>#</th>
					<th>{by}</th>
					<th>流量</th>
				</tr>
			</thead>
			<tbody>
				{limit
					? data && data
							.sort((a, b) => b.period - a.period)
							.slice(0, limit)
							.map((item, index) => {
								return (
									<tr key={item.id}>
										<td>{index + 1}</td>
										{Object.values(item).map((val, i) => (
											<td>{i > 0 ? formatBytes(val) : val}</td>
										))}
									</tr>
								);
							})
					: data && data
							.sort((a, b) => b.period - a.period)
							.map((traffic, index) => {
								return (
									<tr key={index}>
										<td>{index + 1}</td>
										<td>{traffic.period}</td>
										<td>{formatBytes(traffic.amount)}</td>
									</tr>
								);
							})}
			</tbody>
		</Table>
	);
}

export default TrafficTable;
