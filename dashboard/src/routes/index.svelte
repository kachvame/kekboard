<script context="module">
  export const load = async ({ fetch }) => {
    const stats = await fetch('/stats').then((res) => res.json());

    return {
      props: {
        rows: stats.map((stat) => ({
          ...stat,
          id: stat.username,
        })),
      },
    };
  };
</script>

<script>
  import 'carbon-components-svelte/css/g90.css';
  import { DataTable } from 'carbon-components-svelte';

  const headers = [
    {
      key: 'username',
      value: 'Username',
    },
    {
      key: 'count',
      value: 'Count',
    },
  ];

  export let rows;
</script>

<DataTable {headers} {rows} />
