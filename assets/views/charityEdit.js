Vue.directive('init', {
  bind: function(el, binding, vnode) {
    vnode.context[binding.arg] = binding.value;
  }
});

const app = new Vue({
  delimiters: ['${', '}'],
  el: '#app',
  data: {
    errors: [],
    id: null,
    charity: {},
    charityTypes: [],
    itemTypes: [],
    loading: true,
    saving: false,
  },
  mounted() {
    fetch('/api/v1/charity/'+this.id, {
      method: 'get',
    }).then(response => {
      if (response.status !== 200) {
        console.log('error ' + response.status)
        return
      }

      response.json().then(data => {
        this.loading = false
        this.charity = data
      })
    })
  },
  methods:{
    filePicked(event) {
      if (event.target.files.length == 0 ) {
        console.log("no file picked")
        return
      }
      file = event.target.files[0]
      console.log(file)
      reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = e =>{
        this.charity.logo = e.target.result;
      }

    },
    save: function (e) {
      e.preventDefault();

      this.errors = [];
      this.saving = true;

      console.log('saving')

      fetch('/api/v1/charity/'+this.id, {
        method: 'put',
        body: JSON.stringify(this.charity)
      }).then(response => {
        console.log(response)
        window.location.assign('/charity/'+this.id)
      })

    },
  }
});
