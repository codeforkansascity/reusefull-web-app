const app = new Vue({
    delimiters: ['${', '}'],
    el: '#app',
    data: {
        sender: "jdoe@example.com",
        name: "John Doe",
        body: "I found your organization on Re.Use.Full, and would love to arrange for a donation or learn more about your cause.",
        sent: false,
    },
    methods:{
      contact(id) {
        const message = {
            sender: this.sender,
            body: this.body,
            name: this.name,
        }
        fetch(`/api/v1/charity/${id}/contact`, {
          method: 'post',
          body: JSON.stringify(message)
        }).then(
            this.sent = true,
        )
      }
    }
  });
  