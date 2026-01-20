import { Calendar, RefreshCw, User } from 'lucide-react';
import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface NewsItem {
    title: string;
    excerpt: string;
    url: string;
    date: string;
    author: string;
    imageUrl?: string;
}

interface NewsPreviewProps {
    getNews: () => Promise<NewsItem[]>
}

export const NewsPreview: React.FC<NewsPreviewProps> = ({ getNews }) => {
    const { t } = useTranslation();
    const [news, setNews] = useState<NewsItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchNews = async () => {
        setLoading(true);
        setError(null);
        try {
            const items = await getNews();
            setNews(items);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch news');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchNews()
    }, []);

    const openLink = (url: string) => {
        BrowserOpenURL(url);
    };

    return (
        <div className='flex flex-col gap-y-2 max-w-sm'>
            <div className='flex justify-between items-center'>
                <h2 className='text-white text-base font-bold'>
                    {t('Latest Hytale News')}
                </h2>
                <button
                    onClick={fetchNews}
                    disabled={loading}
                    className="rounded-lg hover:text-white disabled:opacity-50 transition-colors">
                    <RefreshCw size={16} className={loading ? 'animate-spin' : ''} />
                </button>
            </div>

            {loading ?
                (<div className="flex items-center justify-center py-4">
                    <div className="text-center">
                        <RefreshCw size={24} className="text-[#FFA845] animate-spin mx-auto mb-2" />
                        <p className="text-white/70 text-xs">{t('Loading news...')}</p>
                    </div>
                </div>)
                : error ? (
                    <div className="flex items-center justify-center py-8">
                        <div className="text-center">
                            <p className="text-red-400 mb-3 text-sm">{error}</p>
                            <button
                                onClick={fetchNews}
                                className="px-4 py-2 bg-white/10 hover:bg-white/20 rounded-lg transition-colors text-sm"
                            >
                                {t('Try Again')}
                            </button>
                        </div>
                    </div>
                ) :
                    (<div className='flex flex-col gap-y-2 glass p-3 rounded-xl'>
                        {news.slice(0, 3).map((item, index) => {
                            return (
                                <button
                                    key={index}
                                    onClick={() => openLink(item.url)}
                                    disabled={loading}
                                    className='flex gap-2 group hover:bg-white/5 p-1.5 rounded-lg transition-all cursor-pointer text-left w-full'
                                >
                                    {/* News Image */}
                                    {item.imageUrl && (
                                        <img
                                            src={item.imageUrl}
                                            alt={item.title}
                                            className='w-20 h-14 object-cover rounded-md flex-shrink-0 group-hover:scale-105 transition-transform'
                                        />
                                    )}
                                    {/* News Content */}
                                    <div className='flex flex-col justify-center min-w-0 pointer-events-none'>
                                        <p className='text-[#FFA845] group-hover:underline text-xs font-medium line-clamp-2 mb-0.5'>
                                            {item.title}
                                        </p>
                                        <div className='flex flex-col text-xs'>
                                            <div className='flex gap-x-1 items-center text-white/50'>
                                                <User size='10' />
                                                <p className='truncate text-xs'>{item.author}</p>
                                            </div>
                                            <div className='flex gap-x-1 items-center text-white/50'>
                                                <Calendar size='10' />
                                                <p className='text-xs'>{item.date}</p>
                                            </div>
                                        </div>
                                    </div>
                                </button>
                            )
                        })}
                        <button
                            onClick={() => openLink("https://hytale.com/news")}
                            className='w-full font-semibold hover:underline cursor-pointer text-[#FFA845] text-xs mt-0.5'>
                            {t('Read more on hytale.com')} →
                        </button>
                    </div>)
            }

        </div>
    );
};
