"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { API_BASE } from "../../api-config";

interface UserStats {
    total_won: number;
    total_lost: number;
    total_refilled: number;
    win_rate: number;
    total_games: number;
}

interface User {
    id: string;
    username: string;
    email: string;
    balance: number;
    tier: string;
    role: string;
    status: string;
    banned_until?: string;
    is_verified: boolean;
    stats?: UserStats;
}

export default function UserDetails() {
    const params = useParams();
    const router = useRouter();
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [balanceAdjustment, setBalanceAdjustment] = useState(0);
    const [banDuration, setBanDuration] = useState("1h");

    const fetchUser = async () => {
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/user?id=${params.id}`, {
                headers: { "Authorization": `Bearer ${token}` }
            });
            if (resp.ok) {
                const data = await resp.json();
                setUser(data);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (params.id) fetchUser();
    }, [params.id]);

    const handleUpdate = async (updates: Partial<User>) => {
        if (!user) return;
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/user`, {
                method: "POST",
                headers: {
                    "Authorization": `Bearer ${token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({ ...user, ...updates })
            });
            if (resp.ok) fetchUser();
        } catch (err) {
            console.error(err);
        }
    };

    const handleAdjustBalance = async () => {
        if (!user || balanceAdjustment === 0) return;
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/user/balance`, {
                method: "POST",
                headers: {
                    "Authorization": `Bearer ${token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    id: user.id,
                    amount: balanceAdjustment,
                    source: "admin_adjustment"
                })
            });
            if (resp.ok) {
                setBalanceAdjustment(0);
                fetchUser();
            }
        } catch (err) {
            console.error(err);
        }
    };

    const handleBan = async (duration: string) => {
        if (!user) return;
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/user/ban`, {
                method: "POST",
                headers: {
                    "Authorization": `Bearer ${token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    id: user.id,
                    duration: duration,
                    is_unban: false
                })
            });
            if (resp.ok) fetchUser();
        } catch (err) {
            console.error(err);
        }
    };

    const handleUnban = async () => {
        if (!user) return;
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/user/ban`, {
                method: "POST",
                headers: {
                    "Authorization": `Bearer ${token}`,
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    id: user.id,
                    is_unban: true
                })
            });
            if (resp.ok) fetchUser();
        } catch (err) {
            console.error(err);
        }
    };

    if (loading) return <div className="text-slate-400">Loading user profile...</div>;
    if (!user) return <div className="text-red-400">User not found</div>;

    const isBanned = user.status === 'banned' || (user.banned_until && new Date(user.banned_until) > new Date());

    return (
        <div className="space-y-6 lg:space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-center gap-3 lg:gap-4">
                    <button
                        onClick={() => router.back()}
                        className="p-1.5 lg:p-2 bg-slate-800 rounded-full hover:bg-slate-700 transition-colors"
                    >
                        <svg className="w-4 h-4 lg:w-5 lg:h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                        </svg>
                    </button>
                    <div>
                        <h1 className="text-xl lg:text-3xl font-bold text-white mb-0.5 lg:mb-1">{user.username}</h1>
                        <p className="text-slate-400 text-[10px] lg:text-sm">User ID: <span className="font-mono">{user.id}</span></p>
                    </div>
                </div>
                <div className="flex gap-2 lg:gap-3">
                    {isBanned ? (
                        <button
                            onClick={handleUnban}
                            className="flex-1 sm:flex-none px-4 py-2 bg-green-600/20 text-green-400 border border-green-600/30 rounded-xl font-bold hover:bg-green-600/30 transition-all text-xs lg:text-base"
                        >
                            UNBAN ACCOUNT
                        </button>
                    ) : (
                        <div className="flex flex-1 sm:flex-none items-center gap-2 bg-slate-900 border border-slate-800 p-1 rounded-xl">
                            <select
                                value={banDuration}
                                onChange={(e) => setBanDuration(e.target.value)}
                                className="flex-1 bg-transparent text-slate-400 text-xs lg:text-sm px-2 lg:px-3 py-1 outline-none min-w-[100px]"
                            >
                                <option value="1h">1 Hour</option>
                                <option value="5h">5 Hours</option>
                                <option value="12h">12 Hours</option>
                                <option value="24h">24 Hours</option>
                                <option value="168h">7 Days</option>
                                <option value="336h">14 Days</option>
                                <option value="720h">30 Days</option>
                                <option value="permanent">Permanent</option>
                            </select>
                            <button
                                onClick={() => handleBan(banDuration)}
                                className="px-4 py-1.5 lg:py-2 bg-red-600 hover:bg-red-500 text-white font-bold rounded-lg transition-all text-xs lg:text-sm"
                            >
                                BAN
                            </button>
                        </div>
                    )}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 lg:gap-8">
                {/* Profile Overview */}
                <div className="lg:col-span-2 space-y-6 lg:space-y-8">
                    <div className="glass p-5 lg:p-8 rounded-2xl grid grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8">
                        <div>
                            <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Account Status</p>
                            <span className={`px-2 py-0.5 lg:py-1 text-[9px] lg:text-[10px] font-bold uppercase rounded border ${isBanned ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-green-500/10 text-green-400 border-green-500/20'}`}>
                                {isBanned ? 'Banned' : 'Active'}
                            </span>
                            {user.banned_until && isBanned && (
                                <p className="text-[9px] lg:text-[10px] text-slate-500 mt-1.5 lg:mt-2 font-mono truncate">Until: {new Date(user.banned_until).toLocaleString()}</p>
                            )}
                        </div>
                        <div>
                            <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Verified</p>
                            <div className="flex items-center gap-2">
                                <span className={`px-2 py-0.5 lg:py-1 text-[9px] lg:text-[10px] font-bold uppercase rounded border ${user.is_verified ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : 'bg-slate-500/10 text-slate-400 border-slate-500/20'}`}>
                                    {user.is_verified ? 'Yes' : 'No'}
                                </span>
                                <button
                                    onClick={() => handleUpdate({ is_verified: !user.is_verified })}
                                    className="px-2 py-0.5 bg-slate-800 hover:bg-slate-700 text-[8px] lg:text-[10px] text-white font-bold rounded border border-slate-700 transition-colors uppercase"
                                >
                                    {user.is_verified ? 'Unverify' : 'Verify'}
                                </button>
                            </div>
                        </div>
                        <div>
                            <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Balance</p>
                            <p className="text-lg lg:text-2xl font-bold text-blue-400 font-mono">${user.balance.toLocaleString()}</p>
                        </div>
                        <div>
                            <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Tier</p>
                            <p className="text-base lg:text-xl font-bold text-white uppercase truncate">{user.tier}</p>
                        </div>
                    </div>

                    <div className="glass p-5 lg:p-8 rounded-2xl">
                        <h2 className="text-lg lg:text-xl font-bold text-white mb-4 lg:mb-6">Performance Statistics</h2>
                        <div className="grid grid-cols-2 lg:grid-cols-4 gap-6 lg:gap-8">
                            <div>
                                <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Total Won</p>
                                <p className="text-lg lg:text-2xl font-bold text-green-400 font-mono">${user.stats?.total_won.toLocaleString() || 0}</p>
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Total Lost</p>
                                <p className="text-lg lg:text-2xl font-bold text-red-400 font-mono">${user.stats?.total_lost.toLocaleString() || 0}</p>
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Total Refilled</p>
                                <p className="text-lg lg:text-2xl font-bold text-blue-400 font-mono">${user.stats?.total_refilled.toLocaleString() || 0}</p>
                            </div>
                            <div>
                                <p className="text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1 lg:mb-2">Win Rate</p>
                                <p className="text-lg lg:text-2xl font-bold text-purple-400 font-mono">{user.stats?.win_rate.toFixed(1) || 0}%</p>
                                <p className="text-[9px] text-slate-500 mt-1 uppercase">Across {user.stats?.total_games || 0} Events</p>
                            </div>
                        </div>
                    </div>

                    <div className="glass p-5 lg:p-8 rounded-2xl">
                        <h2 className="text-lg lg:text-xl font-bold text-white mb-4 lg:mb-6">Account Settings</h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 lg:gap-8">
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1.5 lg:mb-2">User Tier</label>
                                    <select
                                        value={user.tier}
                                        onChange={(e) => handleUpdate({ tier: e.target.value })}
                                        className="w-full px-3 lg:px-4 py-2 lg:py-3 bg-slate-900 border border-slate-800 rounded-xl text-white text-sm lg:text-base outline-none focus:border-blue-500 transition-colors"
                                    >
                                        <option value="VIP Neutral">VIP Neutral</option>
                                        <option value="VIP Bronze">VIP Bronze</option>
                                        <option value="VIP Silver">VIP Silver</option>
                                        <option value="VIP Gold">VIP Gold</option>
                                        <option value="VIP Platinum">VIP Platinum</option>
                                    </select>
                                </div>
                                <div>
                                    <label className="block text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1.5 lg:mb-2">Account Role</label>
                                    <select
                                        value={user.role}
                                        onChange={(e) => handleUpdate({ role: e.target.value })}
                                        className="w-full px-3 lg:px-4 py-2 lg:py-3 bg-slate-900 border border-slate-800 rounded-xl text-white text-sm lg:text-base outline-none focus:border-blue-500 transition-colors"
                                    >
                                        <option value="user">Standard User</option>
                                        <option value="admin">Administrator</option>
                                    </select>
                                </div>
                            </div>
                            <div className="space-y-4 font-mono text-[11px] lg:text-sm text-slate-400">
                                <div className="p-4 bg-slate-900/50 rounded-xl border border-slate-800/50 break-words">
                                    <p className="mb-2">Email: <span className="text-white">{user.email}</span></p>
                                    <p>Username: <span className="text-white">{user.username}</span></p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Balance Controls */}
                <div className="space-y-6 lg:space-y-8">
                    <div className="glass p-5 lg:p-8 rounded-2xl">
                        <h2 className="text-lg lg:text-xl font-bold text-white mb-4 lg:mb-6">Balance Adjustment</h2>
                        <div className="space-y-4">
                            <div>
                                <label className="block text-[10px] lg:text-xs font-semibold text-slate-500 uppercase mb-1.5 lg:mb-2">Amount (Use negative for deduction)</label>
                                <input
                                    type="number"
                                    value={balanceAdjustment}
                                    onChange={(e) => setBalanceAdjustment(parseInt(e.target.value))}
                                    className="w-full px-3 lg:px-4 py-2 lg:py-3 bg-slate-900 border border-slate-800 rounded-xl text-white outline-none focus:border-blue-500 transition-colors font-mono text-sm lg:text-base"
                                    placeholder="e.g. 1000 or -500"
                                />
                            </div>
                            <button
                                onClick={handleAdjustBalance}
                                className="w-full py-3 lg:py-4 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded-xl shadow-lg shadow-blue-900/20 transition-all font-mono text-sm lg:text-base"
                            >
                                EXECUTE ADJUSTMENT
                            </button>
                            <p className="text-[9px] lg:text-[10px] text-slate-500 text-center uppercase tracking-widest">
                                Action will be logged in history
                            </p>
                        </div>
                    </div>

                    <div className="glass p-5 lg:p-8 rounded-2xl bg-gradient-to-br from-red-500/5 to-transparent border-red-500/10">
                        <h2 className="text-lg lg:text-xl font-bold text-red-500 mb-2 lg:mb-4">Danger Zone</h2>
                        <p className="text-slate-500 text-[10px] lg:text-xs mb-4 lg:mb-6">Irreversible administrative actions for this account</p>
                        <button
                            onClick={async () => {
                                if (confirm("DELETE USER? This cannot be undone.")) {
                                    const token = localStorage.getItem("token");
                                    await fetch(`${API_BASE}/admin/user?id=${user.id}`, {
                                        method: "DELETE",
                                        headers: { "Authorization": `Bearer ${token}` }
                                    });
                                    router.push("/user-management");
                                }
                            }}
                            className="w-full py-2.5 lg:py-3 bg-transparent border border-red-500/30 text-red-500 hover:bg-red-500 hover:text-white font-bold rounded-xl transition-all text-xs lg:text-base"
                        >
                            DELETE ACCOUNT
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
