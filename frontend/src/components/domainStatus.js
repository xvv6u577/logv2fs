import { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { doRerender } from "../store/rerender";
import axios from "axios";

const DomainStatus = ({ btnName }) => {

    const [showModal, setShowModal] = useState(false);
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
        <>
            <button
                type="button"
                onClick={() => setShowModal(!showModal)}
                className="w-full sm:w-auto block text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
            >
                <svg
                    aria-hidden="true"
                    className="inline-block mr-1 w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
                </svg>
                {btnName}
            </button>

            {showModal ?
                <div
                    id="add-node-modal"
                    className="overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 w-full md:inset-0 h-modal md:h-full justify-center items-center flex"
                >
                    <div className="relative p-4 w-full max-w-lg h-full md:h-auto">
                        <div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
                            <button
                                type="button"
                                className="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white"
                                onClick={() => setShowModal(!showModal)}
                            >
                                <svg aria-hidden="true" className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
                                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"></path>
                                </svg>
                                <span className="sr-only">Close modal</span>
                            </button>
                            <div className="py-4 px-6 rounded-t border-b dark:border-gray-600">
                                <h3 className="text-base font-semibold text-gray-900 lg:text-xl dark:text-white">
                                    Domain Status
                                </h3>
                            </div>
                            <div className="p-6">
                                <p className="text-sm font-normal text-gray-500 dark:text-gray-400">Days to be Expired!</p>
                                <ul className="my-4 space-y-3">
                                    {domains.map((domain, index) => (
                                        <li key={index} >
                                            <div className="flex items-center p-3 text-base font-bold text-gray-900 bg-gray-50 rounded-lg hover:bg-gray-100 group hover:shadow dark:bg-gray-600 dark:hover:bg-gray-500 dark:text-white">
                                                <span className="d-block flex-1">
                                                    <span className="inline-block w-48">{domain.domain}</span>:{" "}
                                                    <span className="inline-flex justify-center w-20 bg-green-100 text-green-800 text-sm font-medium px-0 py-0.5 rounded dark:bg-green-200 dark:text-green-900">
                                                        {domain.days_to_expire}天
                                                    </span>
                                                </span>
                                                {/* <span className="flex-1 ml-3 whitespace-nowrap">{domain.domain}: {domain.days_to_expire}</span> */}
                                                <span
                                                    onClick={() => {
                                                        var tempDomains = domains.filter(item => item.domain !== domain.domain);
                                                        setDomains(tempDomains);
                                                    }}
                                                    className="cursor-pointer inline-flex items-center justify-center px-2 py-0.5 ml-3 text-xs font-medium text-gray-500 bg-gray-200 rounded dark:bg-gray-700 dark:text-gray-400"
                                                >
                                                    Delete
                                                </span>
                                            </div>
                                        </li>
                                    ))}
                                </ul>
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
                                                        setDomains([...domains, { domain: newdomain, expired_date: "N/A", days_to_expire: "N/A", is_in_uvp: false }]);
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
                : null}

        </>
    );
}

export default DomainStatus;