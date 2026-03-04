"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "../../api-config";

interface SlotRecord {
    id: string;
    username: string;
    user_id: string;
    bet: number;
    lines: number;
    win_amount: number;
    result: number[][];
    winners: number[][]; // [[row, col], ...]
    timestamp: string;
}

export default function SlotHistory() {
    const [history, setHistory] = useState<SlotRecord[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const token = localStorage.getItem("token");
                const resp = await fetch(`${API_BASE}/admin/history/slot`, {
                    headers: { "Authorization": `Bearer ${token}` }
                });
                if (resp.ok) {
                    const data = await resp.json();
                    setHistory(data);
                }
            } catch (err) {
                console.error(err);
            } finally {
                setLoading(false);
            }
        };

        fetchHistory();
    }, []);

    if (loading) return <div className="text-slate-400">Scanning slot logs...</div>;

    return (
        <div className="space-y-6 lg:space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
            <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-white mb-1 lg:mb-2">Slot Machine History</h1>
                <p className="text-slate-400 text-xs lg:text-sm">Real-time logs of slot machine gameplay and outcomes</p>
            </div>

            <div className="glass rounded-2xl overflow-hidden shadow-2xl">
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="bg-slate-900/50 border-b border-slate-800">
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Timestamp</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">User</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Bet Details</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Win</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Net</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider text-center">Matrix</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800 text-sm">
                            {history.map((item) => {
                                const net = item.win_amount - item.bet;
                                return (
                                    <tr key={item.id} className="hover:bg-slate-800/30 transition-colors">
                                        <td className="px-6 py-4 text-slate-500 font-mono whitespace-nowrap">
                                            {item.timestamp ? new Date(item.timestamp).toLocaleString() : 'N/A'}
                                        </td>
                                        <td className="px-6 py-4 font-medium text-white">
                                            <div className="text-white">{item.username || 'Unknown'}</div>
                                            <div className="text-[10px] text-slate-500 font-mono uppercase tracking-tighter truncate w-24" title={item.user_id}>
                                                {item.user_id}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="text-slate-200 font-mono">${(item.bet || 0).toLocaleString()}</div>
                                            <div className="text-[10px] text-slate-500 uppercase">{item.lines || 1} Lines</div>
                                        </td>
                                        <td className="px-6 py-4 font-mono text-green-400 font-bold">
                                            {item.win_amount > 0 ? `+$${item.win_amount.toLocaleString()}` : '-'}
                                        </td>
                                        <td className={`px-6 py-4 font-mono ${net > 0 ? 'text-green-500 font-bold' : net < 0 ? 'text-red-400' : 'text-slate-500'}`}>
                                            {net > 0 ? '+' : ''}{net !== 0 ? net.toLocaleString() : '0'}
                                        </td>
                                        <td className="px-6 py-4">
                                            <div className="flex flex-col gap-1.5 items-center">
                                                {item.result?.map((row, rIdx) => (
                                                    <div key={rIdx} className="flex gap-1.5">
                                                        {row.map((cell, cIdx) => {
                                                            const icons = ["cherry.png", "lemon.png", "orange.png", "plum.png", "watermelon.png", "seven.png"];
                                                            const icon = icons[cell] || "seven.png";
                                                            const isWinner = item.winners?.some(w => w[0] === rIdx && w[1] === cIdx);
                                                            return (
                                                                <div
                                                                    key={cIdx}
                                                                    className={`w-12 h-12 rounded-xl flex items-center justify-center p-1.5 transition-all duration-500 ${isWinner
                                                                            ? 'bg-yellow-500/20 border-2 border-yellow-500/50 shadow-[0_0_15px_rgba(234,179,8,0.2)] scale-110 z-10'
                                                                            : 'bg-slate-900 border border-slate-800'
                                                                        }`}
                                                                >
                                                                    <img
                                                                        src={`/slots/${icon}`}
                                                                        alt={icon}
                                                                        className={`w-full h-full object-contain ${isWinner ? 'drop-shadow-[0_0_8px_rgba(234,179,8,0.8)]' : 'opacity-60'}`}
                                                                    />
                                                                </div>
                                                            );
                                                        })}
                                                    </div>
                                                ))}
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
