import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import axios from "axios";
import { alert, reset, success } from "../store/message";
import Alert from "./alert";
import { doRerender } from "../store/rerender";

function Domain() {
    const [domains, setDomains] = useState([]);
    const [newdomain, updateNewdomain] = useState("");

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);
    const message = useSelector((state) => state.message);
    const rerenderSignal = useSelector((state) => state.rerender);

    useEffect(() => {
        if (message.show === true) {
            setTimeout(() => {
                dispatch(alert({ show: false }));
            }, 5000);
        }
    }, [message, dispatch]);

    useEffect(() => {
        axios
            .get(process.env.REACT_APP_API_HOST + "domaininfo", {
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

        var tempDomainList = {};
        domains.forEach((domain) => {
            if (!domain.is_in_uvp) {
                tempDomainList[domain.domain] = domain.domain;
            }
        })

        axios({
            method: "put",
            url: process.env.REACT_APP_API_HOST + "updatedomaininfo",
            headers: { token: loginState.token },
            data: tempDomainList,
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
            <div className="relative p-4 w-7/12 mx-auto md:h-auto">
                <div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
                    <div className="py-4 px-6 rounded-t border-b dark:border-gray-600">
                        <h3 className="text-base font-semibold text-gray-900 lg:text-xl dark:text-white">
                            Add Domain
                        </h3>
                    </div>
                    <div className="p-6">
                        <table className="table-auto my-4 space-y-3 w-full">
                            <thead>
                                <tr>
                                    <th className="px-4 py-2">Domain</th>
                                    <th className="px-4 py-2">Expired Date</th>
                                    <th className="px-4 py-2">Days to Expire</th>
                                    <th className="px-4 py-2"> </th>
                                </tr>
                            </thead>
                            <tbody>
                                {domains.map((domain) => (
                                    <tr key={domain.domain} className="dark:hover:bg-gray-500" >
                                        <td className="border px-4 py-2">{domain.domain}</td>
                                        <td className="border px-4 py-2">{domain.expired_date}</td>
                                        <td className="border px-4 py-2 text-xl decoration-4">{domain.days_to_expire}</td>
                                        <td className="border px-4 py-2"
                                        >
                                            <button
                                                className="w-full sm:w-auto block mx-auto text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                                                onClick={() => {
                                                    var tempDomains = domains.filter(item => item.domain !== domain.domain);
                                                    setDomains(tempDomains);
                                                }}
                                                disabled={domain.is_in_uvp}
                                            >
                                                Delete
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                        <form className="space-y-6"
                            onSubmit={handleAddDomain}
                        >
                            <div className="relative">
                                <label htmlFor="">Domain:</label>
                                <input
                                    type="text"
                                    onChange={(e) => updateNewdomain(e.target.value.replace(/\s/g, ""))}
                                    value={newdomain}
                                    className="p-4 pl-10 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                    placeholder="New Domain to Add"
                                />
                                <button type="button"
                                    onClick={() => {
                                        if (newdomain.length > 0) {
                                            var tempDomains = domains.filter(item => item.domain === newdomain);
                                            if (tempDomains.length === 0) {
                                                setDomains((prevState) => ([
                                                    ...prevState,
                                                    { domain: newdomain, expired_date: "N/A", days_to_expire: "N/A", is_in_uvp: false },
                                                ]));
                                            }
                                            updateNewdomain("");
                                        } else {
                                            dispatch(alert({ show: true, content: "Domain field should't be left empty." }));
                                        }
                                    }}
                                    className="block text-white absolute right-2.5 bottom-2.5 bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-4 py-2 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                                >
                                    Add Domain
                                </button>
                            </div>
                            <button
                                type="submit"
                                className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                            >
                                Update DomainList
                            </button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default Domain;