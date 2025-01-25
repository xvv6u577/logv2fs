import React, { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { alert, success } from "../store/message";
import { formatBytes } from "../service/service";
import axios from "axios";
import TapToCopied from "./tapToCopied";
import TrafficTable from "./trafficTable";
import { doRerender } from "../store/rerender";

// 提取通用样式
const badgeStyles = {
  base: "inline-flex text-xs font-semibold mr-2 px-2.5 py-0.5 rounded",
  counter: "w-10 bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300",
  name: "w-32 bg-blue-100 text-blue-800 dark:bg-blue-200 dark:text-blue-800",
  status: {
    online: "w-16 bg-green-100 text-green-800 dark:bg-green-200 dark:text-green-900",
    offline: "w-16 bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
  },
  role: {
    admin: "w-16 bg-purple-100 text-purple-800 dark:bg-purple-200 dark:text-purple-900",
    user: "w-16 bg-yellow-100 text-yellow-800 dark:bg-yellow-200 dark:text-yellow-900"
  },
  traffic: "w-24 bg-indigo-100 text-indigo-800 dark:bg-indigo-200 dark:text-indigo-900"
};

// 用户基本信息组件
const UserBasicInfo = ({ index, user, loginState }) => (
  <div>
    <span className={`${badgeStyles.base} ${badgeStyles.counter}`}>
      {index + 1}{"."}
    </span>
    <span className={`${badgeStyles.base} ${badgeStyles.name}`}>
      {user.name}
    </span>
    <span className={`${badgeStyles.base} ${user.status === "plain" ? badgeStyles.status.online : badgeStyles.status.offline}`}>
      {user.status === "plain" ? "online" : "offline"}
    </span>
    <span className={`${badgeStyles.base} ${user.role === "admin" ? badgeStyles.role.admin : badgeStyles.role.user}`}>
      {user.role === "admin" ? "admin" : "user"}
    </span>
    {user.email === loginState.jwt.Email && (
      <span className="bg-pink-100 text-pink-800 text-xs font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-pink-200 dark:text-pink-900">
        Me
      </span>
    )}
  </div>
);

// 流量信息组件
const TrafficInfo = ({ user }) => (
  <span className="flex md:justify-start justify-center items-center w-full md:w-5/12 text-xs">
    {[
      { label: "Today", value: user.daily_logs?.[0]?.traffic },
      { label: "This month", value: user.monthly_logs?.[0]?.traffic },
      { label: "This Year", value: user.yearly_logs?.[0]?.traffic },
      { label: "Used", value: user.used }
    ].map(({ label, value }, index) => (
      <React.Fragment key={index}>
        {label}:{" "}
        <span className={`${badgeStyles.base} ${badgeStyles.traffic}`}>
          {value ? formatBytes(value) : "0 Bytes"}
        </span>
        {" "}
      </React.Fragment>
    ))}
  </span>
);

// 操作按钮组件
const ActionButtons = ({ user, handleOnline, handleOffline, handleDeleteUser, dispatch, rerenderSignal }) => (
  <span className="w-full flex flex-col md:flex-row md:w-1/4">
    <EditUser
      btnName="Edit"
      editUserFunc={() => dispatch(doRerender({ rerender: !rerenderSignal.rerender }))}
      user={user}
    />
    <StatusToggleButton 
      status={user.status} 
      emailId={user.email_as_id}
      onOnline={handleOnline}
      onOffline={handleOffline}
    />
    <ConfirmDelUser
      btnName="Delete"
      deleteUserFunc={() => handleDeleteUser(user.email_as_id)}
    />
  </span>
);

// 状态切换按钮组件
const StatusToggleButton = ({ status, emailId, onOnline, onOffline }) => {
  const isOnline = status === "plain";
  const buttonClass = `w-auto md:w-24 focus:outline-none text-white font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center ${
    isOnline 
      ? "bg-red-700 hover:bg-red-800 dark:bg-red-600 dark:hover:bg-red-700" 
      : "bg-green-700 hover:bg-green-800 dark:bg-green-600 dark:hover:bg-green-700"
  }`;
  
  return (
    <button
      className={buttonClass}
      type="button"
      onClick={() => isOnline ? onOffline(emailId) : onOnline(emailId)}
    >
      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
        <path strokeLinecap="round" strokeLinejoin="round" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
      </svg>
      {isOnline ? "Disable" : "Enable"}
    </button>
  );
};

const UserComp = (props) => {

    const [user, setUser] = useState({});

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);
    const rerenderSignal = useSelector((state) => state.rerender);

    const fetchMore = () => {
        axios
            .get(process.env.REACT_APP_API_HOST + "user/" + props.user.email_as_id, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                setUser(response.data);
            })
            .catch((err) => {
                dispatch(alert({ show: true, content: err.toString() }));
            });
    };

    const handleOnline = (name) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "onlineuser/" + name, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    }

    const handleOffline = (name) => {
        axios
            .get(process.env.REACT_APP_API_HOST + "offlineuser/" + name, {
                headers: { token: loginState.token },
            })
            .then((response) => {
                dispatch(doRerender({ rerender: !rerenderSignal.rerender }));
                dispatch(success({ show: true, content: response.data.message }));
            })
            .catch((err) => {
                if (err.response) {
                    dispatch(alert({ show: true, content: err.response.data }));
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
                    dispatch(alert({ show: true, content: err.response.data }));
                } else {
                    dispatch(alert({ show: true, content: err.toString() }));
                }
            });
    };

    return (
        <>
            <h2 id={`accordion-collapse-heading-${props.index}`}>
                <span className="flex flex-col md:flex-row items-center md:justify-between w-full md:px-5 font-medium text-left border border-b-0 border-gray-200 rounded-t-xl focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-800 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-700 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white">
                    <span className="flex md:justify-start justify-center w-full md:w-1/4">
                        <UserBasicInfo index={props.index} user={props.user} loginState={loginState} />
                    </span>
                    
                    <TrafficInfo user={props.user} />
                    
                    <ActionButtons 
                        user={props.user}
                        handleOnline={handleOnline}
                        handleOffline={handleOffline}
                        handleDeleteUser={handleDeleteUser}
                        dispatch={dispatch}
                        rerenderSignal={rerenderSignal}
                    />
                    
                    <svg
                        onClick={() => {
                            props.update();
                            props.active && fetchMore();
                        }}
                        className={`md:w-1/12 w-10 h-10 shrink-0 dark:hover:bg-gray-600 hover:cursor-pointer ${props.active ? "rotate-180" : "rotate-0"}`}
                        fill="currentColor"
                        viewBox="0 0 20 20"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                </span>
            </h2>
            <div
                id={`accordion-collapse-body-${props.index}`}
                className={`${props.active ? "hidden " : ""}accordion-collapse-body`}
            >
                <div className="w-auto flex flex-col md:w-2/3 md:p-5 mx-auto my-3 px-3 font-light rounded-lg border-4 border-gray-200 dark:border-gray-700 dark:bg-gray-900">
                    <div className="py-1 flex justify-between items-center">
                        <pre className="inline text-sm font-medium text-gray-900 dark:text-white">Email: </pre>
                        <TapToCopied>{user.email_as_id}</TapToCopied>
                    </div>
                    <div className="py-1 flex justify-between items-center">
                        <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">SubUrl:</pre>
                        <TapToCopied>
                            {process.env.REACT_APP_FILE_AND_SUB_URL + "/static/" + user.email_as_id}
                        </TapToCopied>
                    </div>
                    <div className="py-1 flex justify-between items-center">
                        <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">Clash:</pre>
                        <TapToCopied>
                            {process.env.REACT_APP_FILE_AND_SUB_URL + "/clash/" + user.email_as_id + ".yaml"}
                        </TapToCopied>
                    </div>
                    <div className="py-1 flex justify-between items-center">
                        <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">Verge:</pre>
                        <TapToCopied>
                            {process.env.REACT_APP_FILE_AND_SUB_URL + "/verge/" + user.email_as_id}
                        </TapToCopied>
                    </div>
                    <div className="py-1 flex justify-between items-center">
                        <pre className="inline  text-sm font-medium text-gray-900 dark:text-white">Sing-box:</pre>
                        <TapToCopied>
                            {process.env.REACT_APP_FILE_AND_SUB_URL + "/singbox/" + user.email_as_id}
                        </TapToCopied>
                    </div>
                </div>
                <div className="my-4">
                    <div className="pt-2 text-2xl text-center">Monthly Traffic </div>
                    <TrafficTable data={user?.monthly_logs} by="月份" />
                </div>
                <div className="my-4">
                    <div className="pt-2 text-2xl text-center">Daily Traffic</div>
                    <TrafficTable data={user?.daily_logs} by="日期" />
                </div>
            </div>
        </>
    )
}

// function EditUser({ btnName, user, editUserFunc }) {
const EditUser = (props) => {

    const [show, setShow] = useState(false);
    const [role, setRole] = useState(props.user.role);
    const [{ password, name }, setState] = useState({
        password: props.user.password,
        name: props.user.name,
    });

    useEffect(() => {
        setState({
            password: props.user.password,
            name: props.user.name,
        });
        setRole(props.user.role);
    }, [props.user])

    const onChange = (e) => {
        const { name, value } = e.target;
        setState((prevState) => ({ ...prevState, [name]: value }));
    };

    const onChangeRole = (e) => {
        const { value } = e.target;
        setRole(value);
    }

    const dispatch = useDispatch();
    const loginState = useSelector((state) => state.login);

    const handleEditUser = (e) => {
        e.preventDefault();
        setShow(!show);
        console.log(props.user.email_as_id, password, name, role);
        axios({
            method: "post",
            url: process.env.REACT_APP_API_HOST + "edit/" + props.user.email_as_id,
            headers: { token: loginState.token },
            data: {
                "email_as_id": props.user.email_as_id,
                password,
                name,
                role,
            },
        })
            .then((response) => {
                dispatch(success({ show: true, content: "User info updated!" }));
                props.editUserFunc();
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
                className="w-auto md:w-20 focus:outline-none text-white bg-green-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-green-600 dark:hover:bg-green-700 dark:focus:ring-blue-800"
                type="button"
                onClick={() => setShow(!show)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 inline-block mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                {props.btnName}
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
                                <div className="my-4"><span>User Status:
                                    <span className="inline-flex w-32 bg-blue-100 text-blue-800 text-sm mx-2 font-mono font-semibold mr-2 px-2.5 py-0.5 rounded dark:bg-blue-200 dark:text-blue-800" >
                                        {props.user.status}
                                    </span>
                                </span></div>
                                <form className="space-y-6" onSubmit={handleEditUser}>
                                    <div>
                                        <label htmlFor="email" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">Email</label>
                                        <input
                                            type="input"
                                            id="email"
                                            name="email"
                                            placeholder={props.user.email_as_id}
                                            value={props.user.email_as_id}
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
                                            placeholder={name}
                                            value={name}
                                            className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white"
                                        />
                                    </div>
                                    <div>
                                        <label htmlFor="userType" className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-400">User Type</label>
                                        <select
                                            id="userType"
                                            className="block p-2 mb-6 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                            onChange={onChangeRole}
                                            value={role}
                                        >
                                            <option value="normal">Normal</option>
                                            <option value="admin">Admin</option>
                                        </select>
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

const ConfirmDelUser = (props) => {
    const [show, setShow] = useState(false);

    return (
        <>
            <button
                className="w-auto md:w-24 focus:outline-none text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-red-300 font-medium rounded-lg text-sm px-1.5 py-1 m-1 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                type="button"
                onClick={() => setShow(!show)}
            >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1 inline-block" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                {props.btnName}
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
                                        props.deleteUserFunc();
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