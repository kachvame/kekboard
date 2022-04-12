const statsEndpoint = new URL('/stats', process.env.BACKEND_URL).href;

export const get = async () => {
  const stats = await fetch(statsEndpoint).then((res) => res.json());

  return {
    body: stats,
  };
};
