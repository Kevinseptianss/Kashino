"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "../api-config";

interface Stats {
    online_users: number;
    total_users: number;
    total_transactions: number;
}

export default function Dashboard() {
    const [stats, setStats] = useState<Stats | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchStats = async () => {
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/stats`, {
                headers: { "Authorization": `Bearer ${token}` }
            });
            if (resp.ok) {
                const data = await resp.json();
                setStats(data);
            }
        } catch (err) {
            console.error("Failed to fetch dashboard stats", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStats();
        const interval = setInterval(fetchStats, 10000); // Poll every 10s
        return () => clearInterval(interval);
    }, []);

    if (loading) return <div className="text-slate-400">Syncing analytics...</div>;

    return (
        <div className="space-y-6 lg:space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-white mb-1 lg:mb-2">Dashboard Overview</h1>
                <p className="text-slate-400 text-xs lg:text-sm">Real-time system health and user activity</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Stat Card 1 */}
                <div className="glass p-6 rounded-2xl relative overflow-hidden group hover:border-blue-500/50 transition-colors">
                    <div className="absolute top-0 right-0 w-16 h-16 bg-blue-500/10 rounded-bl-[40px] group-hover:bg-blue-500/20 transition-colors flex items-center justify-center">
                        <div className="w-2 h-2 rounded-full bg-blue-500 animate-ping"></div>
                    </div>
                    <p className="text-slate-400 text-sm font-medium mb-1">Users Online</p>
                    <p className="text-4xl font-black text-white">{stats?.online_users || 0}</p>
                    <div className="mt-4 flex items-center text-xs text-green-400">
                        <span className="bg-green-400/10 px-2 py-0.5 rounded mr-2">Live Now</span>
                        Active WS Connections
                    </div>
                </div>

                {/* Stat Card 2 */}
                <div className="glass p-6 rounded-2xl group hover:border-purple-500/50 transition-colors">
                    <p className="text-slate-400 text-sm font-medium mb-1">Total Users</p>
                    <p className="text-4xl font-black text-white">{stats?.total_users || 0}</p>
                    <div className="mt-4 flex items-center text-xs text-blue-400">
                        Registered accounts
                    </div>
                </div>

                {/* Stat Card 3 */}
                <div className="glass p-6 rounded-2xl group hover:border-indigo-500/50 transition-colors">
                    <p className="text-slate-400 text-sm font-medium mb-1">Daily Transactions</p>
                    <p className="text-4xl font-black text-white">{stats?.total_transactions || 0}</p>
                    <div className="mt-4 flex items-center text-xs text-indigo-400">
                        Total spins & hands
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="glass p-8 rounded-2xl">
                    <h2 className="text-xl font-bold text-white mb-6">System Health</h2>
                    <div className="space-y-4">
                        <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-400">Server Status</span>
                                <span className="text-green-400">Operational</span>
                            </div>
                            <div className="w-full h-1 bg-slate-800 rounded-full overflow-hidden">
                                <div className="w-full h-full bg-green-500"></div>
                            </div>
                        </div>
                        <div className="space-y-2">
                            <div className="flex justify-between text-sm">
                                <span className="text-slate-400">Database Latency</span>
                                <span className="text-slate-200">12ms</span>
                            </div>
                            <div className="w-full h-1 bg-slate-800 rounded-full overflow-hidden">
                                <div className="w-[10%] h-full bg-blue-500"></div>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="glass p-8 rounded-2xl flex flex-col items-center justify-center text-center">
                    <div className="w-16 h-16 rounded-full bg-blue-500/20 flex items-center justify-center mb-4">
                        <div className="w-8 h-8 rounded-full border-2 border-blue-500 border-t-transparent animate-spin"></div>
                    </div>
                    <h3 className="text-lg font-bold text-white mb-2">Dynamic Analytics</h3>
                    <p className="text-slate-500 text-sm max-w-[250px]">Advanced charting is being populated with historical data.</p>
                </div>
            </div>
        </div>
    );
}
