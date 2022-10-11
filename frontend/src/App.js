import React from 'react';

function App() {
  return (
      <div className="bg-slate-800 h-screen">
          <div className="flex min-h-full flex-col justify-center py-12 sm:px-6 lg:px-8">
              <div className="sm:mx-auto sm:w-full sm:max-w-md">
                  <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-white">Cosmfaucet</h2>
              </div>

              <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
                  <div className="bg-slate-700 py-8 px-4 shadow sm:rounded-lg sm:px-10">
                      <form className="space-y-6" action="#" method="POST">
                          <div>
                              <label htmlFor="email" className="block text-sm font-medium text-white">
                                  Email address
                              </label>
                              <div className="mt-1">
                                  <input
                                      id="email"
                                      name="email"
                                      type="email"
                                      autoComplete="email"
                                      required
                                      className="block w-full bg-slate-800 text-white appearance-none rounded-md border border-slate-900 px-3 py-2 placeholder-gray-400 shadow-sm focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm"
                                  />
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
                                  />
                              </div>
                          </div>

                          <div>
                              <button
                                  type="submit"
                                  className="flex w-full justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                              >
                                  Run
                              </button>
                          </div>
                      </form>
                  </div>
              </div>
          </div>
      </div>
  );
}

export default App;
