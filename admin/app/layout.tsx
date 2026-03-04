"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import Link from "next/link";
import "./globals.css";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const [isAdmin, setIsAdmin] = useState(false);
  const [username, setUsername] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("token");
    const adminDataString = localStorage.getItem("admin_user");

    if (pathname === "/login") return;

    if (!token || !adminDataString) {
      router.push("/login");
      return;
    }

    try {
      const adminData = JSON.parse(adminDataString);
      if (adminData.role !== "admin") {
        router.push("/login");
        return;
      }
      setIsAdmin(true);
      setUsername(adminData.username);
    } catch (e) {
      router.push("/login");
    }
  }, [pathname, router]);

  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  useEffect(() => {
    // Close sidebar on route change on mobile
    setIsSidebarOpen(false);
  }, [pathname]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("admin_user");
    router.push("/login");
  };

  if (pathname === "/login") {
    return <>{children}</>;
  }

  if (!isAdmin) {
    return <div className="min-h-screen bg-slate-950 flex items-center justify-center text-white">Verifying Access...</div>;
  }

  return (
    <html lang="en">
      <body className="antialiased">
        <div className="flex h-screen bg-slate-950 text-slate-200 overflow-hidden">
          {/* Mobile Header */}
          <header className="lg:hidden fixed top-0 left-0 right-0 h-16 bg-slate-900/80 backdrop-blur-xl border-b border-slate-800 z-50 flex items-center justify-between px-6">
            <h1 className="text-lg font-bold text-white tracking-widest">
              KASHINO <span className="text-blue-500">PRO</span>
            </h1>
            <button
              onClick={() => setIsSidebarOpen(!isSidebarOpen)}
              className="p-2 text-slate-400 hover:text-white transition-colors"
            >
              {isSidebarOpen ? (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              ) : (
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16m-7 6h7" />
                </svg>
              )}
            </button>
          </header>

          {/* Sidebar Overlay */}
          {isSidebarOpen && (
            <div
              className="lg:hidden fixed inset-0 bg-slate-950/60 backdrop-blur-sm z-40"
              onClick={() => setIsSidebarOpen(false)}
            />
          )}

          {/* Sidebar */}
          <aside className={`
            fixed inset-y-0 left-0 z-40 w-64 border-r border-slate-800 bg-slate-900/50 backdrop-blur-xl flex flex-col transition-transform duration-300 lg:relative lg:translate-x-0
            ${isSidebarOpen ? "translate-x-0" : "-translate-x-full"}
          `}>
            <div className="p-6 border-b border-slate-800 hidden lg:block">
              <h1 className="text-xl font-bold text-white tracking-widest">
                KASHINO <span className="text-blue-500">PRO</span>
              </h1>
            </div>

            <nav className="flex-1 p-4 space-y-2 mt-16 lg:mt-0 overflow-y-auto">
              <Link href="/dashboard" className={`flex items-center px-4 py-3 rounded-xl transition-all ${pathname === "/dashboard" ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "hover:bg-slate-800 text-slate-400 hover:text-white"}`}>
                Dashboard
              </Link>
              <Link href="/user-management" className={`flex items-center px-4 py-3 rounded-xl transition-all ${pathname.startsWith("/user-management") ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "hover:bg-slate-800 text-slate-400 hover:text-white"}`}>
                Users CRUD
              </Link>
              <div className="pt-4 pb-2 px-4 text-xs font-semibold text-slate-500 uppercase">History</div>
              <Link href="/history/poker" className={`flex items-center px-4 py-3 rounded-xl transition-all ${pathname === "/history/poker" ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "hover:bg-slate-800 text-slate-400 hover:text-white"}`}>
                Poker History
              </Link>
              <Link href="/history/slot" className={`flex items-center px-4 py-3 rounded-xl transition-all ${pathname === "/history/slot" ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "hover:bg-slate-800 text-slate-400 hover:text-white"}`}>
                Slot History
              </Link>
            </nav>

            <div className="p-4 border-t border-slate-800">
              <div className="flex items-center space-x-3 mb-4 px-2">
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500 flex items-center justify-center text-xs font-bold text-white uppercase shrink-0">
                  {username?.[0] || 'A'}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">{username}</p>
                  <p className="text-xs text-slate-500">Super Admin</p>
                </div>
              </div>
              <button
                onClick={handleLogout}
                className="w-full flex items-center px-4 py-2 text-sm text-red-400 hover:bg-red-500/10 rounded-lg transition-colors border border-transparent hover:border-red-500/20"
              >
                Logout session
              </button>
            </div>
          </aside>

          {/* Main Content */}
          <main className="flex-1 overflow-y-auto p-4 lg:p-8 relative mt-16 lg:mt-0">
            <div className="absolute top-0 right-0 w-1/2 h-1/2 bg-blue-600/5 blur-[120px] rounded-full -z-10"></div>
            <div className="max-w-7xl mx-auto">
              {children}
            </div>
          </main>
        </div>
      </body>
    </html>
  );
}
