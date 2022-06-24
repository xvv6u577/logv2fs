import { Table } from "react-bootstrap";
import { formatBytes } from "../service/service";

function TrafficTable({ data,limit,by }) {
	return (
		<Table striped bordered hover size="sm" variant="dark" className="mx-auto">
			<thead>
				<tr>
					<th>#</th>
					<th>{by}</th>
					<th>Data Used</th>
					<th>By Domain</th>
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
										<td>{item.period}</td>
										<td>{formatBytes(item.amount)}</td>
										<td>{item.used_by_domain && Object.entries(item.used_by_domain).map(([key, value]) => {
											return <span className="d-block" key={key}>{key}:{formatBytes(value)}</span>
										})}
										</td>
									</tr>
								);
							})
					: data && data
							.sort((a, b) => b.period - a.period)
							.map((item, index) => {
								return (
									<tr key={index}>
										<td>{index + 1}</td>
										<td>{item.period}</td>
										<td>{formatBytes(item.amount)}</td>
										<td>{item.amount && Object.entries(item.used_by_domain).map(([key, value]) => {
											return <span className="d-block" key={key}>{key}:{formatBytes(value)}</span>
										})}
										</td>
									</tr>
								);
							})}
			</tbody>
		</Table>
	);
}

export default TrafficTable;
