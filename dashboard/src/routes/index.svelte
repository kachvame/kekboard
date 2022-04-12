<script context="module">
  export const load = async ({ fetch }) => {
    const stats = await fetch('http://localhost:8080/stats').then((res) =>
      res.json(),
    );

    return {
      props: { rows: stats },
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
