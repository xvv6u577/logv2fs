import { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success, reset } from "../store/message";
import { doRerender } from "../store/rerender";
import Alert from "./alert";
import axios from "axios";

const AddNode = () => {

    const [nodes, setnodes] = useState([]);
    const[enable_chatgpt, setChatgptChecked] = useState(false);
    const[enable_subscription, setSubscriptionChecked] = useState(false);
    const initialState = {
        type: "vmess",
        remark: "",
        domain: "",
        ip: "",
        uuid: "",
        path: "",
        sni: ""
	};

    const [{ type, remark, domain, uuid, path, sni, ip }, setState] = useState(initialState);
    const clearState = () => {
        setState({ ...initialState }); 
    };

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);
    const message = useSelector((state) => state.message);
    const rerenderSignal = useSelector((state) => state.rerender);

    useEffect(() => {
        axios
            .get(process.env.REACT_APP_API_HOST + "t7k033", {
                headers: { token: loginState.token },
            })
            .then((response) => {
                setnodes(response.data);
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    }, [rerenderSignal, loginState.token, dispatch]);

    const handleAddNode = (e) => {
        e.preventDefault();
        axios({
            method: "put",
            url: process.env.REACT_APP_API_HOST + "759b0v",
            headers: { token: loginState.token },
            data: nodes,
        })
            .then((response) => {
                dispatch(success({ show: true, content: response.data.message }));
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }))
                clearState();
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    }

    const onChange = (e) => {
        const name = e.target.name;
        const value = e.target.value.replace(/\s/g, "");
        setState((prevState) => ({ ...prevState, [name]: value }));
    };

    useEffect(() => {
		if (message.show === true) {
			setTimeout(() => {
				dispatch(alert({ show: false }));
			}, 5000);
		}
	}, [message, dispatch]);

    return (
        <>
            <Alert message={message.content} type={message.type} shown={message.show} close={() => { dispatch(reset({})); }} />
            <section className="text-gray-400 bg-gray-900 body-font">
                <div className="container px-3 mx-auto">
                    <div className="flex flex-col text-center w-full my-5">
                        <h1 className="sm:text-4xl text-3xl font-medium title-font mb-2 text-white">Node Status</h1>
                        <p className="lg:w-2/3 mx-auto leading-relaxed text-base">Current status of active nodes</p>
                    </div>
                    <div className="px-2 w-full mx-auto overflow-auto shadow-md">
                        <table className="table-auto w-full text-left whitespace-no-wrap">
                            <thead>
                                <tr>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">Type</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">名称</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">域名</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">IP</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">Enable ChatGPT</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">Enable Subscription</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">UUID</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">PATH</th>
                                    <th className="px-4 py-3 title-font tracking-wider font-medium text-white text-sm bg-gray-800">SNI</th>
                                    <th className="w-10 title-font tracking-wider font-medium text-white text-sm bg-gray-800 rounded-tr rounded-br">button</th>
                                </tr>
                            </thead>
                            <tbody>
                                {nodes.map((node, index) => (
                                    <tr key={index + 1000} className="border-b hover:bg-gray-100 dark:hover:bg-gray-700">
                                        <td className="px-4 py-3">{node.type}</td>
                                        <td className="px-4 py-3">{node.remark}</td>
                                        <td className="px-4 py-3">{node.domain}</td>
                                        <td className="px-4 py-3">{node.ip}</td>
                                        <td className="px-4 py-3">{node.enable_chatgpt ? "Yes" : "No"}</td>
                                        <td className="px-4 py-3">{node.enable_subscription ? "Yes" : "No"}</td>
                                        <td className="px-4 py-3">{node.uuid ? node.uuid : "None"}</td>
                                        <td className="px-4 py-3">{node.path ? node.path : "None"}</td>
                                        <td className="px-4 py-3">{node.sni ? node.sni : "None"}</td>
                                        <td className="w-10 text-center">
                                            <span 
                                                onClick={() => {
                                                    // delete node from nodes
                                                    setnodes(nodes.filter((node, i) => i !== index));
                                                }}
                                                className="cursor-pointer inline-flex items-center justify-center px-2 py-0.5 ml-3 text-xs font-medium text-gray-500 bg-gray-200 rounded dark:bg-gray-700 dark:text-gray-400" >
                                                Delete
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                    <form className="" onSubmit={handleAddNode}>
                        <div className="flex w-full mx-auto">
                            <select id="countries"
                                name="type"
                                onChange={onChange}
                                value={type}
                                className="mr-2 my-4 pl-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                            >
                                <option value="vmess">Vmess</option>
                                <option value="vmessws">VmessWS</option>
                                <option value="vmessCDN">VmessCDN</option>
                                <option value="vlessCDN">VLessCDN</option>
                            </select>
                            <input
                                type="text"
                                name="remark"
                                onChange={onChange}
                                value={remark}
                                className="mr-4 my-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="Remark"
                            />
                            <input
                                type="text"
                                name="domain"
                                onChange={onChange}
                                value={domain}
                                className="mr-4 my-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="New domain"
                            />
                            <input
                                type="text"
                                name="ip"
                                onChange={onChange}
                                value={ip}
                                className="mr-4 my-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="New IP"
                            />
                            <input 
                                type="checkbox" 
                                name="enable_chatgpt"
                                onChange={(e)=>setChatgptChecked(e.target.checked)}
                               checked={enable_chatgpt}
                                className="w-4 h-4 m-4 p-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                            <input 
                                type="checkbox" 
                                name="enable_subscription"
                                onChange={(e)=>setSubscriptionChecked(e.target.checked)}
                                checked={enable_subscription}
                                id="default-checkbox" 
                                className="w-4 h-4 m-4 p-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                            <input
                                type="text"
                                name="uuid"
                                onChange={onChange}
                                value={uuid}
                                className="m-4 w-2/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                placeholder="UUID"
                                />
                            <input
                                type="text"
                                name="path"
                                onChange={onChange}
                                value={path}
                                placeholder="Path"
                                className="m-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                />
                            <input
                                type="text"
                                name="sni"
                                onChange={onChange}
                                value={sni}
                                placeholder="SNI"
                                className="m-4 w-1/12 text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                            />
                            <button type="button"
                                onClick={() => {
                                    // append {type, remark, domain, enable_chatgpt, enable_subscription, uuid, path, sni} to nodes
                                    if (domain.length > 0 && remark.length > 0) {
                                        setnodes((prevState) => ([
                                            ...prevState,
                                            {
                                                type,
                                                remark,
                                                domain,
                                                ip,
                                                enable_chatgpt,
                                                enable_subscription,
                                                uuid,
                                                path,
                                                sni
                                            }
                                        ]));
                                        clearState();
                                        setSubscriptionChecked(false);
                                        setChatgptChecked(false);
                                    } else {
                                        dispatch(alert({ show: true, content: "Either the domain or remark field should be left empty." }));
                                    }
                                }}
                                className="w-1/12 px-1 m-4 text-white right-2.5 bottom-2.5 bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                            >
                                Add Domain
                            </button>
                        </div>
                        <button
                            type="submit"
                            className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                        >
                            Update Nodes
                        </button>
                    </form>
                </div>
            </section>
        </>
    )

}

export default AddNode;