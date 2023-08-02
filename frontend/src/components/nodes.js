import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import { alert, reset } from "../store/message";
import Alert from "./alert";
import { doRerender } from "../store/rerender";
import { formatBytes } from "../service/service";

function Nodes() {

    const [nodes, setNodes] = useState([]);
    const [activeTab, setActiveTab] = useState(-1);

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);
    const message = useSelector((state) => state.message);
    const rerenderSignal = useSelector((state) => state.rerender);

    const activateTab = (index) => {
        activeTab === index ? setActiveTab(-1) : setActiveTab(index);
    };

    useEffect(() => {
        if (message.show === true) {
            setTimeout(() => {
                dispatch(alert({ show: false }));
            }, 5000);
        }
    }, [message, dispatch]);

    useEffect(() => {
        axios
            .get(process.env.REACT_APP_API_HOST + "c47kr8", {
                headers: { token: loginState.token },
            })
            .then((response) => {
                setNodes(response.data);
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    }, [loginState, dispatch, rerenderSignal]);

    return (
        <div className="py-3 flex-1 w-full mx-auto">
            <Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }} />
            <div id="accordion-collapse" data-accordion="collapse">
                {nodes && nodes
                // sort by status
                    .sort((a, b) => {
                        if (a.status === "active" && b.status !== "active") {
                            return -1;
                        } else if (a.status !== "active" && b.status === "active") {
                            return 1;
                        } else {
                            return 0;
                        }
                    })
                    .map((node, index) => (
                    <div key={index}>
                        <h2 id={`accordion-collapse-heading-${index}`}>
                            <span className="flex flex-col md:flex-row items-center md:justify-between w-full md:px-5 font-medium text-left border border-b-0 border-gray-200 rounded-t-xl focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-800 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white"
                            >
                                <span className="flex md:justify-start justify-center w-full md:w-2/5">
                                    <div>
                                        <span className="w-10 bg-gray-100 text-gray-800 text-xs font-medium inline-flex items-center px-2.5 py-0.5 rounded mr-2 dark:bg-gray-700 dark:text-gray-300">
                                            {index + 1}{"."}
                                        </span>
                                        <span className={`inline-flex w-50 bg-blue-100 text-blue-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded ${node.status === "active" ? "dark:bg-green-200 dark:text-grebg-green-800" : "dark:bg-blue-200 dark:text-blue-800"}`} >
                                            {node.domain}
                                        </span>
                                        {node.status === "active" ?
                                            (<span className={`inline-flex w-50 bg-blue-100 text-blue-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded ${node.status === "active" ? "dark:bg-green-200 dark:text-grebg-green-800" : "dark:bg-blue-200 dark:text-blue-800"}`} >
                                                {node.remark}
                                            </span>) : null}
                                    </div>
                                </span>
                                <span className="flex md:justify-start justify-center w-full md:w-3/5">
                                    Today:
                                    <span className="inline-flex w-24 bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                                        {formatBytes(node.node_at_current_day.amount)}
                                    </span>
                                    This Month:
                                    <span className="inline-flex w-24 bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                                        {formatBytes(node.node_at_current_month.amount)},
                                    </span>
                                    This Year:
                                    <span className="inline-flex w-24 bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                                        {formatBytes(node.node_at_current_year.amount)}
                                    </span>
                                </span>
                                <svg
                                    onClick={() => {
                                        activateTab(index);
                                    }}
                                    className={`w-10 h-10 shrink-0 dark:hover:bg-gray-600 hover:cursor-pointer ${activeTab !== index ? "rotate-180" : "rotate-0"}`}
                                    fill="currentColor"
                                    viewBox="0 0 20 20"
                                    xmlns="http://www.w3.org/2000/svg">
                                    <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd"></path>
                                </svg>
                            </span>
                        </h2>
                        <div id="accordion-collapse-body-1" className={`${activeTab !== index ? "hidden " : ""}my-4`} aria-labelledby="accordion-collapse-heading-1">
                            <div className="my-4">
                                <div className="pt-2 text-2xl text-center">Monthly Traffic </div>
                                <div className="overflow-x-auto relative shadow-md sm:rounded-lg">
                                    <table className="table-auto w-full text-sm text-left text-gray-500 dark:text-gray-400">
                                        <thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
                                            <tr>
                                                <th scope="col" className="w-1/5 py-4 px-2">#</th>
                                                <th scope="col" className="w-1/5 py-4 px-2">By Month</th>
                                                <th scope="col" className="w-1/5 py-4 px-2">Data Used</th>
                                                <th scope="col" className="w-2/5 py-4 px-2">By Domain</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            { node && node.node_by_month && node.node_by_month
                                                .sort((a, b) => b.period - a.period)
                                                .slice(0, 5)
                                                .map((item, index) => 
                                                 (<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
                                                        <td className="py-4 px-2">{index + 1}</td>
                                                        <td className="py-4 px-2">{item.period}</td>
                                                        <td className="py-4 px-2">{formatBytes(item.amount)}</td>
                                                        <td className="py-4 px-2">{item.user_traffic_at_period && Object.entries(item.user_traffic_at_period).map(([key, value]) => {
                                                            return (
                                                                <div key={key}>
                                                                    <span className="d-block"><span className="inline-block w-44">{key}</span>:{" "}
                                                                        <span className="inline-flex justify-center w-24 bg-green-100 text-green-800 text-sm font-medium mr-2 px-0 py-0.5 rounded dark:bg-green-200 dark:text-green-900">
                                                                            {formatBytes(value)}
                                                                        </span>
                                                                    </span>
                                                                </div>
                                                            )
                                                        })}
                                                        </td>
                                                    </tr>
                                                  )
                                            )}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                            <div className="my-4">
                                <div className="pt-2 text-2xl text-center">Daily Traffic </div>
                                <div className="overflow-x-auto relative shadow-md sm:rounded-lg">
                                    <table className="table-auto w-full text-sm text-left text-gray-500 dark:text-gray-400">
                                        <thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
                                            <tr>
                                                <th scope="col" className="w-1/5 py-4 px-2">#</th>
                                                <th scope="col" className="w-1/5 py-4 px-2">By Day</th>
                                                <th scope="col" className="w-1/5 py-4 px-2">Data Used</th>
                                                <th scope="col" className="w-2/5 py-4 px-2">By Domain</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            { node && node.node_by_day && node.node_by_day
                                                .sort((a, b) => b.period - a.period)
                                                .slice(0, 14)
                                                .map((item, index) => 
                                                 (<tr key={index} className="bg-white border-b dark:bg-gray-900 dark:border-gray-700">
                                                        <td className="py-4 px-2">{index + 1}</td>
                                                        <td className="py-4 px-2">{item.period}</td>
                                                        <td className="py-4 px-2">{formatBytes(item.amount)}</td>
                                                        <td className="py-4 px-2">{item.user_traffic_at_period && Object.entries(item.user_traffic_at_period).map(([key, value]) => {
                                                            return (
                                                                <div key={key}>
                                                                    <span className="d-block"><span className="inline-block w-44">{key}</span>:{" "}
                                                                        <span className="inline-flex justify-center w-24 bg-green-100 text-green-800 text-sm font-medium mr-2 px-0 py-0.5 rounded dark:bg-green-200 dark:text-green-900">
                                                                            {formatBytes(value)}
                                                                        </span>
                                                                    </span>
                                                                </div>
                                                            )
                                                        })}
                                                        </td>
                                                    </tr>
                                                  )
                                            )}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>))}
            </div>
        </div>);
}

export default Nodes;