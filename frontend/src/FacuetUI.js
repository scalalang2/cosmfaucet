import React, {useEffect} from "react";
import SelectChainUI from "./SelectChainUI";
import clsx from "clsx";

export default function FaucetUI({ isSending, chains, selectedChain, handleSelectChain, address, onAddressChanged, handleSubmit }) {
    return (
        <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
            <div className="bg-slate-700 py-8 px-4 shadow sm:rounded-lg sm:px-10">
                <form className="space-y-6" action="#" onSubmit={handleSubmit}>
                    <div>
                        <label htmlFor="email" className="block text-sm font-medium text-white">
                            Select Chains
                        </label>
                        <div className="mt-1">
                            <SelectChainUI chains={chains} selectedChain={selectedChain} handleSelectChain={handleSelectChain} />
                        </div>
                    </div>

                    <div>
                        <label htmlFor="address" className="block text-sm font-medium text-white">
                            Wallet Address
                        </label>
                        <div className="mt-1">
                            <input
                                id="address"
                                name="address"
                                type="text"
                                required
                                className="block w-full bg-slate-800 text-white appearance-none rounded-md border border-slate-900 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm"
                                value={address}
                                onChange={e => onAddressChanged(e.target.value)}
                            />
                        </div>
                    </div>

                    <div>
                        <button
                            type="submit"
                            className={clsx("flex w-full justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2",{
                                "opacity-50 cursor-not-allowed": isSending
                            })}
                            disabled={isSending}
                        >
                            Run
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
