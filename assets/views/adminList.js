const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
  },
  methods:{
    approve: function (id) {
      fetch(`/api/v1/charity/${id}/approve`, {
        method: 'put'
      }).then(response => {
        location.reload();
      })
    },
    deny: id => {
      fetch(`/api/v1/charity/${id}/deny`, {
        method: 'put'
      }).then(response => {
        location.reload();
      })
    }
  }
});
