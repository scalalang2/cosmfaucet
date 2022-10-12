import React from "react";
import { ExclamationTriangleIcon, XCircleIcon, CheckCircleIcon } from '@heroicons/react/20/solid'
import clsx from "clsx";

export default function AlertView({ message, type }){
    return (
        <div className={clsx({
            "rounded-md bg-red-50 p-4": type === "error",
            "rounded-md bg-yellow-50 p-4": type === "warning",
            "rounded-md bg-green-50 p-4": type === "success",
        })}>
            <div className="flex">
                <div className="flex-shrink-0">
                    {type === "warning" && (<ExclamationTriangleIcon className="h-5 w-5 text-yellow-400" aria-hidden="true" />)}
                    {type === "error" && (<XCircleIcon className="h-5 w-5 text-red-400" aria-hidden="true" />)}
                    {type === "success" && (<XCircleIcon className="h-5 w-5 text-green-400" aria-hidden="true" />)}
                </div>
                <div className="ml-3">
                    <h3 className={clsx({
                        "text-sm font-medium text-yellow-800": type === "warning",
                        "text-sm font-medium text-red-800": type === "error",
                        "text-sm font-medium text-green-800": type === "success",
                    })}>Attention needed</h3>
                    <div className={
                        clsx({
                            "mt-2 text-sm text-yellow-700": type === "warning",
                            "mt-2 text-sm text-red-700": type === "error",
                            "mt-2 text-sm text-green-700": type === "success",
                        })
                    }>
                        <p>
                            {message}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    )
}