import { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { formatBytes } from "../service/service";
import axios from "axios";
import TapToCopied from "./tapToCopied";
import TrafficTable from "./trafficTable";
import { doRerender } from "../store/rerender";


const UserComp = (props) => {

    const [user, setUser] = useState({});

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);
    const rerenderSignal = useSelector((state) => state.rerender);

    const fetchMore = () => {
        axios
            .get(process.env.REACT_APP_API_HOST + "user/" + props.user.email, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                setUser(response.data);
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
                console.log(err.toString());
            });
    };

    const handleOnline = (name) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "takeuseronline/" + name, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    };

    const handleOffline = (name) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "takeuseroffline/" + name, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data.error }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    const handleDeleteUser = (name) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "deluser/" + name, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data.error }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    const handleDisableNode = ({ email, node }) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "disanodeperusr", {
                params: { email, node },
                headers: { token: loginState.token },
            })
            .then((response) => {
                // dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                fetchMore()
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data.error }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    const handleEnableNode = ({ email, node }) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "enanodeperusr", {
                params: { email, node },
                headers: { token: loginState.token },
            })
            .then((response) => {
                // dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                fetchMore()
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data.error }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    return (
        <>
            <h2 id={`accordion-collapse-heading-${props.index}`} >
                <span className="flex flex-col md:flex-row items-center md:justify-between w-full md:px-5 font-medium text-left border border-b-0 border-gray-200 rounded-t-xl focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-800 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white"
                >
                    <span className="flex md:justify-start justify-center w-full md:w-1/3">
                        <div>
                            <span className="bg-gray-100 text-gray-800 text-xs font-medium inline-flex items-center px-2.5 py-0.5 rounded mr-2 dark:bg-gray-700 dark:text-gray-300">
                                {props.index + 1}{"."}
                            </span>
                            <span className="bg-blue-100 text-blue-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-blue-200 dark:text-blue-800" >
                                {props.user.name}
                            </span>
                            {props.user.status === "plain" ? (
                                <span className="bg-green-100 text-green-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-green-200 dark:text-green-900">
                                    online
                                </span>
                            ) : (
                                <span className="bg-gray-100 text-gray-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-gray-700 dark:text-gray-300">
                                    offline
                                </span>
                            )}
                            {props.user.role === "admin" ? (
                                <span className="bg-purple-100 text-purple-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-purple-200 dark:text-purple-900">
                                    admin
                                </span>
                            ) : (
                                <span className="bg-yellow-100 text-yellow-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-yellow-200 dark:text-yellow-900">
                                    user
                                </span>
                            )}
                            {props.user.email === loginState.jwt.Email && (
                                <span className="bg-pink-100 text-pink-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-pink-200 dark:text-pink-900">
                                    Me
                                </span>
                            )}
                        </div>
                    </span>
                    <span className="flex md:justify-start justify-center items-center w-full  md:w-1/3 text-xs"> Today:{" "}
                        <span className="bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                            {formatBytes(props.user.used_by_current_day.amount)}
                        </span>
                        {" "}
                        This month:{" "}
                        <span className="bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                            {formatBytes(props.user.used_by_current_month.amount)}
                        </span>
                        {" "} Used:{" "}
                        <span className="bg-indigo-100 text-indigo-800 text-sm font-medium mr-2 px-2.5 py-0.5 rounded dark:bg-indigo-200 dark:text-indigo-900">
                            {formatBytes(props.user.used)}
                        </span>
                    </span>
                    <span className="w-full flex flex-col md:flex-row md:w-1/3">
                        <EditUser
                            btnName="Edit"
                            editUserFunc={() =>
                                dispatch(doRerender({ rerender: !rerenderSignal.rerender }))
                            }
                            user={props.user}
                        />
                        {props.user.status === "plain" ? (
                            <button
                                className="focus:outline-none text-white bg-red-700 hover:bg-red-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
                                // className="focus:outline-none text-white font-medium rounded-lg text-sm px-2.5 py-1 m-1 text-center dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800 w-full md:w-24 md:h-10"
                                type="button"
                                onClick={() => handleOffline(props.user.email)}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                                </svg>
                                Disable
                            </button>

                        ) : (
                            <button
                                className="focus:outline-none text-white bg-green-700 hover:bg-green-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800"
                                // className="focus:outline-none text-white font-medium rounded-lg text-sm px-2.5 py-1 m-1 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800 w-full md:w-24 md:h-10"
                                type="button"
                                onClick={() => handleOnline(props.user.email)}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                                    <path strokeLinecap="round" strokeLinejoin="round" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                                </svg>
                                Enable
                            </button>
                        )}
                        <ConfirmDelUser
                            btnName="Delete"
                            deleteUserFunc={() => handleDeleteUser(props.user.email)}
                        />
                    </span>
                    <svg
                        onClick={() => {
                            props.update();
                            props.active && fetchMore();
                        }}
                        className={`w-10 h-10 shrink-0 dark:hover:bg-gray-600 hover:cursor-pointer ${props.active ? "rotate-180" : "rotate-0"}`}
                        fill="currentColor"
                        viewBox="0 0 20 20"
                        xmlns="http://www.w3.org/2000/svg">
                        <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd"></path>
                    </svg>
                </span>
            </h2>
            <div
                id={`accordion-collapse-body-${props.index}`}
                className={`${props.active ? "hidden " : ""}accordion-collapse-body`}
            >
                <div className="flex flex-col md:flex-row md:p-5 font-light border border-b-0 border-gray-200 dark:border-gray-700 dark:bg-gray-900">
                    <div className="flex flex-col content-between rounded-lg border-4 border-neutral-100 mx-auto my-3 md:p-3 md:w-2/5" >
                        <div className="py-1 flex justify-between items-center">
                            <pre className="inline text-sm font-medium text-gray-900 dark:text-white">Email: </pre>
                            <TapToCopied>{user.email}</TapToCopied>
                        </div>
                        <div className="py-1 flex justify-between items-center">
                            <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">Path:</pre>
                            <TapToCopied>{user.path}</TapToCopied>
                        </div>
                        <div className="py-1 flex justify-between items-center">
                            <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">UUID:</pre>
                            <TapToCopied>{user.uuid}</TapToCopied>
                        </div>
                        <div className="py-1 flex justify-between items-center">
                            <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">SubUrl:</pre>
                            <TapToCopied>
                                {process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + user.email}
                            </TapToCopied>
                        </div>
                        <div className="py-1 flex justify-between items-center">
                            <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">Clash:</pre>
                            <TapToCopied>
                                {process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + user.email + ".yaml"}
                            </TapToCopied>
                        </div>
                    </div>

                    <div className="flex flex-col content-between rounded-lg border-4 border-neutral-100 md:mx-auto md:my-3 md:p-3 md:w-1/3">
                        {user && user.node_in_use_status &&
                            Object.entries(user.node_in_use_status).map(
                                ([key, value]) => (
                                    <div className="flex flex-row justify-between" key={key}>
                                        <span className="flex items-center text-sm">{key}</span>
                                        {value ? (
                                            <button
                                                type="button"
                                                style={{width: "100px", height: "40px"}}
                                                className="inline-flex justify-center text-white bg-red-700 hover:bg-red-800 focus:ring-4 focus:outline-none focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
                                                onClick={() =>
                                                    handleDisableNode({
                                                        email: user.email,
                                                        node: key,
                                                    })
                                                }
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" viewBox="0 0 20 20" fill="currentColor">
                                                    <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                                                </svg>
                                                Disable
                                            </button>
                                        ) : (
                                            <button
                                                type="button"
                                                style={{width: "100px", height: "40px"}}
                                                className="inline-flex justify-center text-white bg-green-700 hover:bg-green-800 focus:ring-4 focus:outline-none focus:ring-green-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-green-800"
                                                onClick={() =>
                                                    handleEnableNode({
                                                        email: user.email,
                                                        node: key,
                                                    })
                                                }
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" viewBox="0 0 20 20" fill="currentColor">
                                                    <path d="M10 2a5 5 0 00-5 5v2a2 2 0 00-2 2v5a2 2 0 002 2h10a2 2 0 002-2v-5a2 2 0 00-2-2H7V7a3 3 0 015.905-.75 1 1 0 001.937-.5A5.002 5.002 0 0010 2z" />
                                                </svg>
                                                Enable
                                            </button>
                                        )}
                                    </div>
                                )
                            )}
                    </div>
                </div>
                <div className="my-4">
                    <div className="pt-2 text-2xl text-center">Monthly Traffic </div>
                    <TrafficTable data={user.traffic_by_month} by="月份" />
                </div>
                <div className="my-4">
                    <div className="pt-2 text-2xl text-center">Daily Traffic</div>
                    <TrafficTable data={user.traffic_by_day} by="日期" />
                </div>
            </div>
        </>
    )
}

function EditUser({ btnName, user, editUserFunc }) {
    const [show, setShow] = useState(false);

    const [{ used, password, name, role, credit }, setState] = useState({
        used: user.used,
        password: user.password,
        name: user.name,
        role: user.role,
        credit: user.credit,
    });

    const onChange = (e) => {
        const { name, value } = e.target;
        setState((prevState) => ({ ...prevState, [name]: value }));
    };

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);

    const handleEditUser = (e) => {
        e.preventDefault();
        setShow(!show);
        axios({
            method: "post",
            url: process.env.REACT_APP_API_HOST + "edit/" + user.email,
            headers: { token: loginState.token },
            data: {
                role,
                email: user.email,
                password,
                name,
                used: parseInt(used),
                credit: parseInt(credit),
            },
        })
            .then((response) => {
                dispatch(success({ show: true, content: "user info updated!" }));
                editUserFunc();
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data.error }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    return (
        <>
            <button
                className="focus:outline-none text-white bg-green-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-blue-800"
                // className="focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                type="button"
                onClick={() => setShow(!show)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                {btnName}
            </button>

            {show ?
                <div
                    id="add-user-modal"
                    className="overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 w-full md:inset-0 h-modal md:h-full justify-center items-center flex" aria-modal="true"
                    role="dialog"
                >
                    <div className="relative p-4 w-full max-w-md h-full md:h-auto">
                        <div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
                            <button
                                type="button"
                                className="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white"
                                onClick={() => setShow(!show)}
                            >
                                <svg aria-hidden="true" className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"></path></svg>
                                <span className="sr-only">Close modal</span>
                            </button>
                            <div className="py-6 px-6 lg:px-8">
                                <h3 className="mb-4 text-xl font-medium text-gray-900 dark:text-white">Edit User</h3>
                                <div><span>User Status: <b>{user.status}</b></span></div>
                                <div><span>UUID: <b>{user.uuid}</b></span></div>
                                <div><span>Path:  <b>{user.path}</b></span></div>
                                <form className="space-y-6" onSubmit={handleEditUser}>
                                    <div>
                                        <label htmlFor="email" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Email</label>
                                        <input
                                            type="input"
                                            id="email"
                                            name="email"
                                            placeholder={user.email}
                                            value={user.email}
                                            className="bg-gray-5ˀ0 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                            disabled />
                                    </div>
                                    <div>
                                        <label htmlFor="password" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Password (6+ characters)</label>
                                        <input
                                            type="password"
                                            name="password"
                                            id="password"
                                            value={password}
                                            onChange={onChange}
                                            placeholder="••••••••"
                                            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                        />
                                    </div>
                                    <div>
                                        <label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Name</label>
                                        <input
                                            type="input"
                                            id="name"
                                            name="name"
                                            onChange={onChange}
                                            placeholder={user.name}
                                            value={name}
                                            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                        />
                                    </div>
                                    <div>
                                        <label htmlFor="userType" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-400">User Type</label>
                                        <select
                                            id="userType"
                                            className="block p-2 mb-6 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                            onChange={onChange}
                                            value={role}
                                        >
                                            <option value="normal">Normal</option>
                                            <option value="admin">Admin</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">已用流量</label>
                                        <input
                                            type="number"
                                            name="used"
                                            onChange={onChange}
                                            placeholder={user.used}
                                            value={used}
                                            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                        />
                                    </div>
                                    <div>
                                        <label htmlFor="path" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">每月限额</label>
                                        <input
                                            type="number"
                                            name="credit"
                                            onChange={onChange}
                                            placeholder={user.credit}
                                            value={credit}
                                            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                        />
                                    </div>

                                    <button type="submit" className="w-full text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800">Submit</button>

                                </form>
                            </div>
                        </div>
                    </div>
                </div>
                : null}
        </>
    );
}

function ConfirmDelUser({ btnName, deleteUserFunc }) {
    const [show, setShow] = useState(false);

    return (
        <>
            <button
                className="focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                // className="w-full sm:w-auto  flex text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-2.5 py-2.5 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                type="button"
                onClick={() => setShow(!show)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1 inline-block" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                {btnName}
            </button>
            {show ?
                <div className="overflow-y-auto overflow-x-hidden fixed top-0 right-0 left-0 z-50 md:inset-0 h-modal md:h-full justify-center items-center flex" >
                    <div className="relative p-4 w-full max-w-md h-full md:h-auto">
                        <div className="relative bg-white rounded-lg shadow dark:bg-gray-700">
                            <button type="button" className="absolute top-3 right-2.5 text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm p-1.5 ml-auto inline-flex items-center dark:hover:bg-gray-800 dark:hover:text-white"
                                onClick={() => setShow(!show)}
                                data-modal-toggle="popup-modal">
                                <svg aria-hidden="true" className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd"></path></svg>
                                <span className="sr-only">Close modal</span>
                            </button>
                            <div className="p-6 text-center">
                                <svg aria-hidden="true" className="mx-auto mb-4 w-14 h-14 text-gray-400 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                                <h3 className="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">Are you sure you want to delete this user?</h3>
                                <button
                                    type="button"
                                    onClick={() => {
                                        setShow(!show);
                                        deleteUserFunc();
                                    }}
                                    className="text-white bg-red-600 hover:bg-red-800 focus:ring-4 focus:outline-none focus:ring-red-300 dark:focus:ring-red-800 font-medium rounded-lg text-sm inline-flex items-center px-5 py-2.5 text-center mr-2">
                                    Yes, I'm sure
                                </button>
                                <button
                                    type="button"
                                    onClick={() => setShow(!show)}
                                    className="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-gray-200 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2.5 hover:text-gray-900 focus:z-10 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
                                >
                                    No, cancel
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
                : null}
        </>
    );
}

export default UserComp;