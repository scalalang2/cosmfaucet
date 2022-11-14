import React, {useEffect, useState} from 'react';
import FaucetUI from "./FacuetUI";
import AlertView from "./AlertView";

function App() {
    const [chains, setChains] = useState([]);
    const [error, setError] = useState(null);
    const [alert, setAlert] = useState(null);
    const [success, setSuccess] = useState(null);
    const [isSending, setIsSending] = useState(false);
    const [address, setAddress] = useState("");
    const [selected, setSelected] = useState(null);

    const loadChains = async () => {
        return fetch("/api/v1/faucet/chains")
            .then(response => response.json())
    }

    const handleSubmit = async (event) => {
        event.preventDefault();
        setIsSending(true);
        setSuccess(null);
        setAlert(null);
        let payload = {
            address: address,
            chainId: selected.chainId,
        };
        const response = await fetch("/api/v1/faucet/give_me", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(payload)
        });
        let res_json = await response.json();

        setIsSending(false);
        response.status === 200 ? setSuccess("your request is sent, it may takes some time to proceed") : setAlert(res_json.message);
        response.status === 200 ? setAlert(null) : setSuccess(null);
        console.log(response);
    }

    useEffect(() => {
        loadChains().then(res => {
            setChains(res.chains)
            setSelected(res.chains[0])
        }).catch(err => {
            console.error(err);
            setError("server is not available");
        })
    }, [])

    if(error != null) {
        return (
            <div className="bg-slate-800 h-screen">
                <div className="flex min-h-full flex-col justify-center py-12 sm:px-6 lg:px-8">
                    <div className="sm:mx-auto sm:w-full sm:max-w-md">
                        <AlertView message={error} type="error"/>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="bg-slate-800 h-screen">
            <div className="flex min-h-full flex-col justify-center py-12 sm:px-6 lg:px-8">
                <div className="sm:mx-auto sm:w-full sm:max-w-md">
                    { error && <AlertView message={error} type="error"/> }
                    { alert && <AlertView message={alert} type="warning"/> }
                    { success && <AlertView message={success} type="success"/> }
                    { isSending && <AlertView message={"Transaction is being sent to the server"} type="warning"/> }
                  <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-white">Cosmfaucet</h2>
                </div>
                { chains.length === 0 && (<div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
                    <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
                        <div className="bg-slate-700 py-8 px-4 shadow sm:rounded-lg sm:px-10">
                            <div className="space-y-6 text-white">
                                No chains available
                            </div>
                        </div>
                    </div>
                </div>) }
                <FaucetUI
                    isSending={isSending}
                    chains={chains}
                    selectedChain={selected}
                    handleSelectChain={setSelected}
                    address={address}
                    onAddressChanged={setAddress}
                    handleSubmit={handleSubmit}
                />
            </div>
        </div>
    );
}

export default App;
