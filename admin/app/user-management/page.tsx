"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "../api-config";
import Link from "next/link";
import { useRouter } from "next/navigation";

interface UserStats {
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
    stats?: UserStats;
}

export default function UserManagement() {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(true);
    const router = useRouter();

    const fetchUsers = async () => {
        try {
            const token = localStorage.getItem("token");
            const resp = await fetch(`${API_BASE}/admin/users`, {
                headers: { "Authorization": `Bearer ${token}` }
            });
            if (resp.ok) {
                const data = await resp.json();
                setUsers(data);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers();
    }, []);

    if (loading) return <div className="text-slate-400 p-8">Loading user database...</div>;

    return (
        <div className="space-y-6 lg:space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div className="flex justify-between items-end">
                <div>
                    <h1 className="text-2xl lg:text-3xl font-bold text-white mb-1 lg:mb-2">User Management</h1>
                    <p className="text-slate-400 text-xs lg:text-sm">Review activity and modify user accounts</p>
                </div>
            </div>

            <div className="glass rounded-2xl overflow-hidden shadow-2xl">
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="bg-slate-900/50 border-b border-slate-800">
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Username</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Email</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Balance</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Status</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider text-right">Win Rate</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800">
                            {users.map((user) => {
                                const isBanned = user.status === 'banned' || (user.banned_until && new Date(user.banned_until) > new Date());
                                const winRate = user.stats?.win_rate || 0;
                                const isHighWinRate = winRate >= 50;

                                return (
                                    <tr
                                        key={user.id}
                                        onClick={() => router.push(`/user-management/${user.id}`)}
                                        className="hover:bg-slate-800/50 transition-colors cursor-pointer group"
                                    >
                                        <td className="px-6 py-4">
                                            <div className="font-medium text-white group-hover:text-blue-400 transition-colors">{user.username}</div>
                                            <div className="text-[10px] text-slate-500 font-mono uppercase tracking-tighter truncate w-24" title={user.id}>
                                                {user.id}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4 text-slate-400">{user.email}</td>
                                        <td className="px-6 py-4 text-blue-400 font-mono">${user.balance.toLocaleString()}</td>
                                        <td className="px-6 py-4">
                                            <span className={`px-2 py-1 text-[10px] font-bold uppercase rounded border ${isBanned ? 'bg-red-500/10 text-red-400 border-red-500/20' : 'bg-green-500/10 text-green-400 border-green-500/20'}`}>
                                                {isBanned ? 'Banned' : 'Active'}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 text-right leading-tight">
                                            <div className={`text-base font-bold font-mono ${isHighWinRate ? 'text-red-400' : 'text-green-400'}`}>
                                                {winRate.toFixed(1)}%
                                            </div>
                                            <div className="text-[9px] text-slate-500 uppercase">
                                                {user.stats?.total_games || 0} Games
                                            </div>
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
