import Image from "next/image";

export default function Home() {
  return (
    <div className="min-h-screen bg-[#050510] text-zinc-100 font-sans selection:bg-purple-500 selection:text-white">
      {/* Hero Section */}
      <nav className="fixed top-0 w-full z-50 bg-black/50 backdrop-blur-lg border-b border-white/10">
        <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
          <div className="text-2xl font-bold bg-gradient-to-r from-purple-400 to-amber-400 bg-clip-text text-transparent">
            KASHINO
          </div>
          <div className="hidden md:flex gap-8 text-sm font-medium uppercase tracking-widest text-zinc-400">
            <a href="#" className="hover:text-white transition-colors">Beranda</a>
            <a href="#fitur" className="hover:text-white transition-colors">Fitur</a>
            <a href="#tentang" className="hover:text-white transition-colors">Tentang Kami</a>
          </div>
          <button className="px-6 py-2 bg-gradient-to-r from-purple-600 to-indigo-600 rounded-full text-sm font-bold hover:scale-105 transition-transform shadow-lg shadow-purple-500/20 uppercase tracking-wider">
            Main Sekarang
          </button>
        </div>
      </nav>

      <main>
        {/* Hero Section */}
        <section className="relative h-screen flex items-center justify-center overflow-hidden">
          <div className="absolute inset-0 z-0">
            <Image
              src="/hero_bg.png"
              alt="Background Kasino"
              fill
              className="object-cover opacity-40 scale-105"
              priority
            />
            <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#050510]/80 to-[#050510]" />
          </div>

          <div className="relative z-10 text-center px-6 max-w-4xl">
            <h1 className="text-5xl md:text-8xl font-black mb-6 tracking-tighter leading-tight">
              RASAKAN <span className="bg-gradient-to-r from-amber-200 via-amber-400 to-amber-600 bg-clip-text text-transparent italic">SENSASINYA</span>
            </h1>
            <p className="text-xl md:text-2xl text-zinc-400 mb-10 max-w-2xl mx-auto leading-relaxed">
              Mainkan Game Kasino Favoritmu Tanpa Uang Sungguhan. Murni Hiburan, Murni Kesenangan, Seratus Persen Aman.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <button className="px-10 py-5 bg-white text-black font-black rounded-lg hover:bg-zinc-200 transition-colors uppercase tracking-widest text-sm shadow-xl shadow-white/10">
                Mulai Petualangan
              </button>
              <button className="px-10 py-5 border border-white/20 font-black rounded-lg hover:bg-white/5 transition-colors uppercase tracking-widest text-sm backdrop-blur-md">
                Lihat Game
              </button>
            </div>
          </div>

          <div className="absolute bottom-10 left-1/2 -translate-x-1/2 animate-bounce opacity-20">
            <div className="w-1 h-12 bg-white rounded-full" />
          </div>
        </section>

        {/* Features Section */}
        <section id="fitur" className="py-32 px-6 relative overflow-hidden">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-24">
              <h2 className="text-sm font-bold text-purple-500 uppercase tracking-[0.3em] mb-4">Kenapa Memilih Kami?</h2>
              <p className="text-4xl md:text-6xl font-black tracking-tight">PENGALAMAN KASINO SOSIAL TERBAIK</p>
            </div>

            <div className="grid md:grid-cols-3 gap-8">
              {[
                {
                  title: "TANPA UANG SUNGGUHAN",
                  desc: "Nikmati semua sensasi menang tanpa risiko finansial sedikitpun.",
                  icon: "💎",
                  img: "/slots_feature.png"
                },
                {
                  title: "MODERN & PREMIUM",
                  desc: "Desain antarmuka berkelas dunia yang membuat Anda merasa seperti di VIP Lounge.",
                  icon: "✨",
                  img: "/hero_bg.png"
                },
                {
                  title: "KEADILAN TERJAMIN",
                  desc: "Algoritma kami dirancang untuk memberikan pengalaman bermain yang adil dan menyenangkan.",
                  icon: "⚖️",
                  img: "/poker_feature.png"
                }
              ].map((feature, idx) => (
                <div key={idx} className="group relative bg-zinc-900/50 border border-white/5 rounded-3xl p-8 hover:bg-zinc-800/50 transition-all duration-500">
                  <div className="relative h-48 w-full mb-8 overflow-hidden rounded-2xl">
                    <Image src={feature.img} alt={feature.title} fill className="object-cover group-hover:scale-110 transition-transform duration-700 opacity-60" />
                  </div>
                  <div className="text-4xl mb-4">{feature.icon}</div>
                  <h3 className="text-xl font-black mb-4 tracking-wider">{feature.title}</h3>
                  <p className="text-zinc-500 leading-relaxed font-medium">{feature.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* About Section */}
        <section id="tentang" className="py-32 bg-white text-black rounded-[4rem] mx-4">
          <div className="max-w-7xl mx-auto px-6 grid md:grid-cols-2 gap-20 items-center">
            <div>
              <h2 className="text-sm font-bold text-zinc-400 uppercase tracking-[0.3em] mb-4">Tentang Kashino</h2>
              <p className="text-5xl md:text-7xl font-black tracking-tighter mb-8 leading-none">
                MISI KAMI ADALAH <span className="text-zinc-400">HIBURAN TANPA BATAS.</span>
              </p>
              <p className="text-xl text-zinc-600 mb-10 leading-relaxed font-medium">
                Kashino didirikan untuk memberikan ruang bagi para penggemar kasino di Indonesia untuk bersenang-senang dan bersosialisasi tanpa perlu khawatir tentang aspek negatif perjudian. Di sini, kemenangan Anda adalah kebanggaan, bukan keuntungan materi.
              </p>
              <div className="grid grid-cols-2 gap-8 border-t border-black/10 pt-10">
                <div>
                  <div className="text-4xl font-black mb-2">100%</div>
                  <div className="text-xs font-bold text-zinc-500 uppercase tracking-widest">Aman & Gratis</div>
                </div>
                <div>
                  <div className="text-4xl font-black mb-2">24/7</div>
                  <div className="text-xs font-bold text-zinc-500 uppercase tracking-widest">Ketersediaan Game</div>
                </div>
              </div>
            </div>
            <div className="relative aspect-square">
              <Image
                src="/poker_feature.png"
                alt="Poker Premium"
                fill
                className="object-cover rounded-[3rem] shadow-2xl"
              />
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="py-40 text-center px-6">
          <div className="max-w-4xl mx-auto bg-gradient-to-br from-indigo-900 to-purple-900 p-16 md:p-32 rounded-[3.5rem] relative overflow-hidden shadow-2xl shadow-purple-950/20">
            <div className="relative z-10">
              <h2 className="text-4xl md:text-7xl font-black mb-10 tracking-tight leading-none">SIAP UNTUK MENANG BESAR?</h2>
              <button className="px-12 py-6 bg-white text-black font-black rounded-full hover:scale-110 transition-transform shadow-xl shadow-black/20 uppercase tracking-[0.2em] text-sm">
                Gabung Sekarang Juga
              </button>
            </div>
            <div className="absolute top-0 left-0 w-full h-full bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] opacity-10" />
          </div>
        </section>
      </main>

      <footer className="py-20 border-t border-white/5 opacity-40">
        <div className="max-w-7xl mx-auto px-6 text-center">
          <div className="text-sm font-bold tracking-[0.5em] mb-8">KASHINO SOCIAL CLUB</div>
          <p className="text-xs max-w-lg mx-auto leading-loose font-medium">
            Permainan ini ditujukan untuk audiens dewasa hanya untuk tujuan hiburan. Keberhasilan dalam permainan kasino sosial tidak menjamin keberhasilan masa depan dalam perjudian uang sungguhan.
          </p>
          <div className="mt-12 text-[10px] font-bold tracking-widest uppercase">
            &copy; 2024 KASHINO. Hak Cipta Dilindungi.
          </div>
        </div>
      </footer>
    </div>
  );
}
