"use client";

import { useEffect, useState } from "react";
import { API_BASE } from "../../api-config";

interface Card {
    suit: string;
    value: string;
}

interface PokerRecord {
    id: string;
    room_id: string;
    hand_id: string;
    event: string;
    username: string;
    amount: number;
    pot: number;
    cards: Card[];
    community: Card[];
    details: string;
    timestamp: string;
}

const CardDisplay = ({ card }: { card: Card }) => {
    if (!card) return null;
    const valueMap: Record<string, string> = {
        "10": "10",
        "jack": "jack",
        "queen": "queen",
        "king": "king",
        "ace": "ace"
    };
    const val = valueMap[card.value] || card.value;
    const fileName = `${val}_of_${card.suit}.png`;

    return (
        <div className="w-8 h-12 bg-white rounded border border-slate-300 shadow-sm flex-shrink-0 relative overflow-hidden group">
            <img
                src={`/cards/${fileName}`}
                alt={`${card.value} of ${card.suit}`}
                className="w-full h-full object-cover"
                onError={(e) => {
                    const target = e.target as HTMLImageElement;
                    if (!target.src.includes("back.png")) {
                        target.src = "/cards/back.png";
                    }
                }}
            />
            <div className="absolute inset-0 bg-black/0 group-hover:bg-black/5 transition-colors" />
        </div>
    );
};

export default function PokerHistory() {
    const [history, setHistory] = useState<PokerRecord[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchHistory = async () => {
            try {
                const token = localStorage.getItem("token");
                const resp = await fetch(`${API_BASE}/admin/history/poker`, {
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

    if (loading) return <div className="text-slate-400 p-8">Reading hand logs...</div>;

    return (
        <div className="space-y-6 lg:space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700 p-4 lg:p-8">
            <div>
                <h1 className="text-2xl lg:text-3xl font-bold text-white mb-1 lg:mb-2">Network Poker Logs</h1>
                <p className="text-slate-400 text-xs lg:text-sm">Comprehensive table of poker events and hand data</p>
            </div>

            <div className="glass rounded-2xl overflow-hidden shadow-2xl border border-slate-800">
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="bg-slate-900/80 border-b border-slate-800">
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Time / User</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Event</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Hand</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Board</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Amount</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Pot</th>
                                <th className="px-6 py-4 text-xs font-semibold text-slate-500 uppercase tracking-wider">Hand Ranking</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800 text-sm">
                            {history.map((item) => (
                                <tr key={item.id} className="hover:bg-slate-800/20 transition-colors group">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="text-slate-500 font-mono text-[11px]">
                                            {item.timestamp ? new Date(item.timestamp).toLocaleString() : 'N/A'}
                                        </div>
                                        <div className="font-bold text-slate-100 mt-1 uppercase tracking-tight">
                                            {item.username || '-'}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span className={`px-2.5 py-1 text-[10px] font-black uppercase rounded-lg border ${item.event === 'win' ? 'bg-green-500/10 text-green-400 border-green-500/20 shadow-[0_0_10px_rgba(34,197,94,0.1)]' :
                                            item.event === 'fold' ? 'bg-red-500/10 text-red-400 border-red-500/20' :
                                                'bg-blue-500/10 text-blue-400 border-blue-500/20'
                                            }`}>
                                            {item.event}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex gap-1 min-w-[70px]">
                                            {item.cards?.map((card, idx) => (
                                                <CardDisplay key={idx} card={card} />
                                            ))}
                                            {!item.cards && <span className="text-slate-600 italic text-xs">No cards</span>}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex gap-1">
                                            {item.community?.map((card, idx) => (
                                                <CardDisplay key={idx} card={card} />
                                            ))}
                                            {(!item.community || item.community.length === 0) && <span className="text-slate-700 italic text-[10px] uppercase">Pre-Flop</span>}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-slate-200 font-mono font-bold">
                                        {typeof item.amount === 'number' && item.amount !== 0 ? (
                                            <span className={item.amount > 0 ? 'text-green-400' : 'text-slate-300'}>
                                                {item.amount > 0 ? '+' : ''}${item.amount.toLocaleString()}
                                            </span>
                                        ) : '-'}
                                    </td>
                                    <td className="px-6 py-4 text-slate-400 font-mono">
                                        ${(item.pot || 0).toLocaleString()}
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="text-slate-400 text-xs italic group-hover:text-slate-200 transition-colors">
                                            {item.details || '-'}
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
