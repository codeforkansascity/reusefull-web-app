Vue.directive('init', {
  bind: function (el, binding, vnode) {
    vnode.context[binding.arg] = binding.value;
  },
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
  computed: {
    formPhone: {
      get() {
        /*
            On change, filter any non-nums, and 
            return formatted phone num for display.
        */
        if (this.charity.phone) {
          const clean = this.charity.phone.replace(/[^\d]/g, '');
          const cleanLen = clean.length;

          if (cleanLen < 4) {
            return clean;
          } else if (cleanLen < 7) {
            return `(${clean.slice(0, 3)}) ${clean.slice(3)}`;
          } else {
            return `(${clean.slice(0, 3)}) ${clean.slice(3, 6)}-${clean.slice(
              6,
              10
            )}`;
          }
        }
      },
      set(num) {
        // update phone in data with new input
        const clean = num.replace(/[^\d]/g, '');
        if (clean.length <= 10) {
          this.charity.phone = clean;
        }
      },
    },
  },
  mounted() {
    fetch('/api/v1/charity/' + this.id, {
      method: 'get',
    }).then((response) => {
      if (response.status !== 200) {
        console.log('error ' + response.status);
        return;
      }

      response.json().then((data) => {
        this.loading = false;
        this.charity = data;
        console.log(data)
      });
    });
  },
  methods: {
    filePicked(event) {
      if (event.target.files.length == 0) {
        console.log('no file picked');
        return;
      }
      file = event.target.files[0];
      console.log(file);
      reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = (e) => {
        this.charity.logo = e.target.result;
      };
    },
    save: function (e) {
      e.preventDefault();

      this.errors = [];
      this.saving = true;

      // Removes any remaining chars in phone num.
      this.charity.phone = this.charity.phone.replace(/[^\d]/g, '');

      console.log('saving');

      fetch('/api/v1/charity/' + this.id, {
        method: 'put',
        body: JSON.stringify(this.charity),
      }).then((response) => {
        console.log(response);
        window.location.assign('/charity/' + this.id);
      });
    },
  },
});
