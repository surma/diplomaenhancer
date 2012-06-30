define(function() {
	var $password = $('#password');
	var $newurl = $('#newurl');
	var $addbtn = $('#add');
	var $activate = $('#activate');
	var $table = $('#table');
	var $body = $($table.find('tbody'));
	var $notificationarea = $('#notificationarea');
	var $templates = $('.template');
	var $successtpl = $($templates.filter('.alert-success'));
	var $errortpl = $($templates.filter('.alert-error'));
	var $rowtpl = $($templates.filter('tr'));

	var timeToLive = function(obj, time) {
		setTimeout(function() {
			$(obj).remove();
		}, time*1000);
	}

	var obj = {
		setState: function(state, password) {
			var oldstate = state != 'active'?'checked':false;
			$.ajax({
				'url': '/api/state',
				'data': state,
				'type': 'POST',
				'headers': {
					'X-DiplomaEnhancer-Token': password,
				},
				'statusCode': {
					400: function() {
						this.showError('Invalid state');
						$activate.attr('disabled');

					}.bind(this),
					401: function() {
						this.showError('Wrong password');
						$activate.attr('disabled');
					}.bind(this),
					204: function() {
						this.showSuccess('DiplomaEnhancer is now '+state);
						$activate.attr('disabled');
					}.bind(this),
				},
				error: function() {
					$activate.attr('checked', oldstate);
				}
			});
		},
		showSuccess: function(msg) {
			var $success = $($successtpl.clone().removeClass('template'));
			$($success.find('.text')).text(msg);
			$success.appendTo($notificationarea);
			timeToLive($success, 5);
		},
		showError: function(msg) {
			if(msg == null || msg == undefined) {
				msg = 'Some error occured';
			}
			var $error = $($errortpl.clone().removeClass('template'));
			$($error.find('.text')).text(msg);
			$error.appendTo($notificationarea);
			timeToLive($error, 5);
		},
		showMessage: function(msg) {
			if(msg instanceof Array) {
				for(m in msg) {
					this.showMessage(m);
				}
				return;
			}

			if(msg.type == 'success') {
				this.showSuccess(msg.text);
			} else {
				this.showError(msg.text);
			}
		},
		clearTable: function() {
			$table.find('tbody tr').each(function() {
				if(!$(this).hasClass('template')) {
					$(this).remove();
				}
			});
		},
		addRow: function(uuid, url) {
			var $row = $($rowtpl.clone().removeClass('template').attr('id', uuid));
			$row.find('button').click(function(){
				var password = $password.attr('value');
				this.unblock(uuid, password);
			}.bind(this));
			$row.children('.url').text(url);
			$row.appendTo($body);
		},
		removeRow: function(uuid) {
			$("#"+uuid).remove();
		},
		updateTable: function() {
			this.clearTable();
			$.ajax({
				'url': '/api/',
				'dataType': 'json',
				'success': function(data) {
					for(var uuid in data) {
						for(var urlidx in data[uuid].Hostnames) {
							var url = data[uuid].Hostnames[urlidx];
							this.addRow(uuid, url);
						}
					}
				}.bind(this),
				'error': function() {
					this.showMessage({
						'type': 'error',
						'text': 'Could not load table',
					});
				}.bind(this),
			});
		},
		block: function(url) {
			var entry = {
				IP: '127.0.0.1',
				Hostnames: [
					url,
				],
			}
			$.ajax({
				'url': '/api/',
				'data': JSON.stringify(entry),
				'type': 'POST',
				'statusCode': {
					201: function(uuid) {
						this.addRow(uuid, url);
						this.showMessage({
							'type': 'success',
							'text': 'Blocked '+url,
						});
						$newurl.attr('value', '');
					}.bind(this),
				},
				'error': function() {
					this.showMessage({
						'type': 'error',
					})
				}
			});
		},
		unblock: function(uuid, password) {
			$.ajax({
				'url': '/api/'+uuid,
				'type': 'DELETE',
				'headers': {
					'X-DiplomaEnhancer-Token': password,
				},
				'statusCode': {
					400: function() {
						this.showMessage({
							'type': 'error',
							'text': 'Invalid request',
						});
					}.bind(this),
					401: function() {
						this.showMessage({
							'type': 'error',
							'text': 'Wrong password',
						});
						return true;
					}.bind(this),
					204: function() {
						this.showMessage({
							'type': 'success',
							'text': 'Unblocked',
						});
						this.removeRow(uuid);
					}.bind(this),
				},
			});
		},
	};

	$addbtn.click(function() {
		var url = $newurl.attr('value');
		var password = $password.attr('value');
		obj.block(url, password);
	});

	$activate.click(function() {
		console.log('drin');
		$activate.attr('disabled');
		var state = $activate.attr('checked')?'active':'inactive';
		var password = $password.attr('value');
		obj.setState(state, password);
	})
	return obj;
});
