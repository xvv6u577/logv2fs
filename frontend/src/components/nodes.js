import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import { alert, reset, success } from "../store/message";
import Alert from "./alert";
import { doRerender } from "../store/rerender";
import { formatBytes } from "../service/service";

function Nodes() {

    const [nodes, setNodes] = useState([]);
    const [activeTab, setActiveTab] = useState(-1);

    const [domains, setDomains] = useState([]);
    const [newdomain, updateNewdomain] = useState("");
    const [newRemark, updateNewRemark] = useState("");


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

        axios
            .get(process.env.REACT_APP_API_HOST + "681p32", {
                headers: { token: loginState.token },
            })
            .then((response) => {
                setDomains(response.data);
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    }, [loginState, dispatch, rerenderSignal]);

    const handleAddDomain = (e) => {
        e.preventDefault();

        axios({
            method: "put",
            url: process.env.REACT_APP_API_HOST + "g7302b",
            headers: { token: loginState.token },
            data: domains,
        })
            .then((response) => {
                dispatch(success({ show: true, content: response.data.message }));
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }))
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    }

    return (
        <div className="py-3 flex-1 w-full mx-auto">
            <Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }} />
            <h2 className="text-2xl font-semibold leading-tight">
                Work Domains Status
                <p className="inline-block px-5 text-sm font-normal text-gray-500 dark:text-gray-400">Days to be Expired!</p>
            </h2>
            <div className="container px-5 py-5 mx-auto">
                <div className="flex flex-wrap -m-4">
                    {
                        domains.map((domain, index) => (
                            <div className="p-4 lg:w-1/3" key={index} >
                                <div className="h-full bg-gray-800 bg-opacity-40 px-8 pt-10 pb-10 rounded-lg overflow-hidden text-center relative hover:bg-gray-700">
                                    <button className="absolute top-0 right-0 p-2"
                                        onClick={() => {
                                            var tempDomains = domains.filter(item => item.domain !== domain.domain);
                                            setDomains(tempDomains);
                                        }}
                                    >

                                        <svg xmlns="http://www.w3.org/2000/svg"
                                            className="h-5 w-5 fill-current text-gray-300 hover:text-red-600"
                                            viewBox="0 0 50 50"
                                        >
                                            <path d="M 7.71875 6.28125 L 6.28125 7.71875 L 23.5625 25 L 6.28125 42.28125 L 7.71875 43.71875 L 25 26.4375 L 42.28125 43.71875 L 43.71875 42.28125 L 26.4375 25 L 43.71875 7.71875 L 42.28125 6.28125 L 25 23.5625 Z"></path>
                                        </svg>
                                    </button>
                                    <h2 className="w-full tracking-widest text-sm title-font font-medium text-sky-400 mb-1">
                                        {domain.domain}
                                        <span className="w-30 px-2 inline-flex justify-center bg-green-100 text-green-800 text-sm font-medium py-0.5 rounded dark:bg-green-200 dark:text-green-900">
                                            {domain.remark}
                                        </span>
                                    </h2>
                                    <h1 className="title-font sm:text-4xl text-2xl font-medium text-white mb-3 py-3">{domain.days_to_expire}天</h1>
                                    <p className="leading-relaxed mb-3">到期时间: {domain.expired_date}</p>
                                </div>
                            </div>
                        ))
                    }
                </div>
                <form className="space-y-6 py-5 px-72" onSubmit={handleAddDomain} >
                    <div className="relative">
                        <span className="inline-block w-1/3 pr-3">
                            <label htmlFor="" className="block">Domain:</label>
                            <input
                                type="text"
                                onChange={(e) => updateNewdomain(e.target.value.replace(/\s/g, ""))}
                                value={newdomain}
                                className="w-full p-4 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="New Domain"
                            />
                        </span>
                        <span className="inline-block w-1/3 pr-3">
                            <label htmlFor="" className="block">Remark:</label>
                            <input
                                type="text"
                                onChange={(e) => updateNewRemark(e.target.value.replace(/\s/g, ""))}
                                value={newRemark}
                                className="w-full p-4 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="Remark of Domain"
                            />
                        </span>
                        <span className="inline-block w-1/3">
                            <button type="button"
                                onClick={() => {
                                    if (newdomain.length > 0 && newRemark.length > 0) {
                                        var tempDomains = domains.filter(item => item.domain === newdomain);
                                        if (tempDomains.length === 0) {
                                            setDomains([...domains, { domain: newdomain, remark: newRemark }]);
                                        }
                                        updateNewdomain("");
                                        updateNewRemark("");
                                    } else {
                                        dispatch(alert({ show: true, content: "Domain field should't be left empty." }));
                                    }
                                }}
                                className="mx-auto w-full block text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-4 py-4 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                            >
                                Add Domain
                            </button>
                        </span>
                    </div>
                    <button
                        type="submit"
                        className="px-4 py-4 mx-auto text-center text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm sm:w-auto dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                    >
                        Update Domains
                    </button>
                </form>
            </div>
            <h2 className="text-2xl font-semibold leading-tight py-3">Active Nodes</h2>
            <div className="p-6" id="accordion-collapse" data-accordion="collapse">
                {nodes && nodes
                    // sort by status
                    // .sort((a, b) => {
                    //     if (a.status === "active" && b.status !== "active") {
                    //         return -1;
                    //     } else if (a.status !== "active" && b.status === "active") {
                    //         return 1;
                    //     } else {
                    //         return 0;
                    //     }
                    // })
                    // remove nodes with status "inactive"
                    .filter((node) => node.status === "active")
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
                                                {node && node.node_by_month && node.node_by_month
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
                                                {node && node.node_by_day && node.node_by_day
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